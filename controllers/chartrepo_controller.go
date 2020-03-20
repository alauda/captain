/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Masterminds/vcs"
	"github.com/alauda/captain/pkg/helm"
	"github.com/alauda/captain/pkg/util"
	"github.com/alauda/helm-crds/pkg/apis/app/v1beta1"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/prometheus/common/log"
	"gopkg.in/src-d/go-git.v4"
	githttp "gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"helm.sh/helm/pkg/repo"
	"io"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"net/http"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/yaml"
	"strings"
	"sync"
	"time"
)

// ChartRepoReconciler reconciles a ChartRepo object
type ChartRepoReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	// we only want to watch one namespace ,this is the easy way...
	Namespace string

	// store latest commit for every vcs repo to compare. Use sycn map to avoid concurrent op(just in case)
	CommitMap sync.Map
}

// +kubebuilder:rbac:groups=alauda.io.alauda.io,resources=chartrepoes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=alauda.io.alauda.io,resources=chartrepoes/status,verbs=get;update;patch

func (r *ChartRepoReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("chartrepo", req.NamespacedName)

	if req.NamespacedName.Namespace != r.Namespace {
		return ctrl.Result{}, nil
	}

	// your logic here
	var cr v1beta1.ChartRepo
	if err := r.Get(ctx, req.NamespacedName, &cr); err != nil {
		log.Error(err, "unable to fetch chartrepo")
		return ctrl.Result{}, ignoreNotFound(err)
	}

	if !r.isReadyForResync(&cr) {
		log.Info("not ready for resync")
		return ctrl.Result{}, nil
	}

	log.Info("resync chartrepo")

	if err := r.syncChartRepo(&cr, ctx); err != nil {
		log.Error(err, "sync chartrepo failed")
		return ctrl.Result{}, r.updateChartRepoStatus(ctx, &cr, v1beta1.ChartRepoFailed, err.Error())
	} else {
		return ctrl.Result{}, r.updateChartRepoStatus(ctx, &cr, v1beta1.ChartRepoSynced, "")
	}

}

func (r *ChartRepoReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.ChartRepo{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 3}).
		Complete(r)
}

func isNotFound(err error) bool {
	return apierrs.IsNotFound(err)
}

func ignoreNotFound(err error) error {
	if isNotFound(err) {
		return nil
	}
	return err
}

// syncChartRepo sync chartrepo. If this is a VCS repo, built charts from the source
// and build/refresh the helm repo. If this is a helm repo already, update the charts if need to
func (r *ChartRepoReconciler) syncChartRepo(cr *v1beta1.ChartRepo, ctx context.Context) error {
	log := r.Log.WithValues("chartrepo", cr.GetName())
	log.Info("chartrepo type is ", "type", cr.Spec.Type)

	if cr.Spec.Type == "Git" {
		flag, err := r.buildChartRepoFromGit(ctx, cr)
		if err != nil {
			return err
		}
		if err := r.updateChartRepoURL(ctx, cr); err != nil {
			return err
		}
		if flag == false {
			return nil
		}
	}

	if cr.Spec.Type == "SVN" {
		flag, err := r.buildChartRepoFromSvn(ctx, cr)
		if err != nil {
			return err
		}

		if err := r.updateChartRepoURL(ctx, cr); err != nil {
			return err
		}

		if flag == false {
			return nil
		}

	}

	// only helm repo need to refresh charts now
	return r.syncCharts(cr, ctx)
}

func (r *ChartRepoReconciler) buildChartRepoFromSvn(ctx context.Context, cr *v1beta1.ChartRepo) (bool, error) {
	log := r.Log.WithValues("chartrepo", cr.GetName())
	dir := "/tmp/svn-temp/" + cr.Name

	if cr.Spec.Source == nil {
		return false, errors.New("no source specified for svn repo")
	}

	data, err := r.GetSecretData(cr, ctx)
	if err != nil {
		return false, err
	}
	if data == nil {
		data = map[string][]byte{}
	}

	s, err := vcs.NewSvnRepo(cr.Spec.Source.URL, dir)
	if err != nil {
		log.Error(err, "init svn repo error")
		return false, err
	}
	s.Username = string(data["username"])
	s.Password = string(data["password"])

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		log.Info("svn source already cloned, check for updates")
		if err := s.Update(); err != nil {
			return false, err
		}

	} else {
		log.Info("svn clone source", "dir", dir, "url", cr.Spec.Source.URL)
		if err := s.Get(); err != nil {
			log.Error(err, "checkout svn repo error")
			return false, err
		}

	}

	version, err := s.Version()
	if err != nil {
		log.Error(err, "get svn rev head error")
		return false, err
	}

	log.Info("svn head is", "ref", version)
	result, ok := r.CommitMap.Load(cr.Name)
	if ok {
		if result.(string) == version {
			log.Info("svn repo is already update to date")
			return false, nil
		}
	}

	log.Info("build charts from svn source")
	if err := helm.SouceToChartRepo(cr.Name, dir, cr.Spec.Source.Path); err != nil {
		return false, err
	}

	r.CommitMap.Store(cr.Name, version)
	return true, nil

}

// the flag indicated wether  we need to rebuild the charts
func (r *ChartRepoReconciler) buildChartRepoFromGit(ctx context.Context, cr *v1beta1.ChartRepo) (bool, error) {
	log := r.Log.WithValues("chartrepo", cr.GetName())
	dir := "/tmp/git-temp/" + cr.Name

	if cr.Spec.Source == nil {
		return false, errors.New("no source specified for git repo")
	}

	data, err := r.GetSecretData(cr, ctx)
	if err != nil {
		return false, err
	}
	if data == nil {
		data = map[string][]byte{}
	}

	var gr *git.Repository

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		log.Info("source already cloned, check for updates")
		g, err := git.PlainOpen(dir)
		if err != nil {
			return false, err
		}
		w, err := g.Worktree()
		if err != nil {
			return false, err
		}
		err = w.Pull(&git.PullOptions{
			RemoteName: "origin",
			Auth: &githttp.BasicAuth{
				Username: string(data["username"]),
				Password: string(data["password"]),
			},
		})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			return false, err
		}
		gr = g
	} else {
		log.Info("git clone source", "dir", dir, "url", cr.Spec.Source.URL)

		g, err := git.PlainClone(dir, false, &git.CloneOptions{
			Auth: &githttp.BasicAuth{
				Username: string(data["username"]),
				Password: string(data["password"]),
			},
			URL:      cr.Spec.Source.URL,
			Progress: os.Stdout,
		})
		if err != nil {
			return false, err
		}
		gr = g
	}

	ref, err := gr.Head()
	if err != nil {
		log.Error(err, "get rev head error")
		return false, err
	}

	// store for feature compare
	latest := ref.Hash().String()
	log.Info("git head is", "ref", latest)

	result, ok := r.CommitMap.Load(cr.Name)
	if ok {
		if result.(string) == latest {
			log.Info("git source is already update to date")
			return false, nil
		}
	}

	log.Info("build charts from git source")
	if err := helm.SouceToChartRepo(cr.Name, dir, cr.Spec.Source.Path); err != nil {
		return false, err
	}

	r.CommitMap.Store(cr.Name, latest)
	return true, nil

}

// GetSecretData get secret auth data for chartrepo/git/svn...
func (r *ChartRepoReconciler) GetSecretData(cr *v1beta1.ChartRepo, ctx context.Context) (map[string][]byte, error) {
	if cr.Spec.Secret != nil {
		ns := cr.Spec.Secret.Namespace
		if ns == "" {
			ns = cr.Namespace
		}

		key := client.ObjectKey{Namespace: ns, Name: cr.Spec.Secret.Name}

		var secret corev1.Secret
		if err := r.Get(ctx, key, &secret); err != nil {
			// If created by UI, the secret is owned by the ChartRepo object, so it may be not found for now
			// but we can wait and try again.
			if apierrs.IsNotFound(err) {
				log.Error("secret not found for now, wait and try again")
				time.Sleep(3 * time.Second)
				if err := r.Get(ctx, key, &secret); err != nil {
					return nil, err
				} else {
					data := secret.Data
					return data, nil
				}

			}
			return nil, err
		}

		data := secret.Data
		return data, nil

	}
	//TODO: return empty dict
	return nil, nil
}

// DownloadIndexFile fetches the index from a repository.
func (r *ChartRepoReconciler) GetIndex(cr *v1beta1.ChartRepo, ctx context.Context) (*repo.IndexFile, error) {
	var username string
	var password string

	if cr.Spec.Secret != nil {
		ns := cr.Spec.Secret.Namespace
		if ns == "" {
			ns = cr.Namespace
		}

		key := client.ObjectKey{Namespace: ns, Name: cr.Spec.Secret.Name}

		var secret corev1.Secret
		if err := r.Get(ctx, key, &secret); err != nil {
			return nil, err
		}

		data := secret.Data
		username = string(data["username"])
		password = string(data["password"])

	}

	link := strings.TrimSuffix(cr.Spec.URL, "/") + "/index.yaml"
	c := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", link, nil)

	if username != "" && password != "" {
		req.SetBasicAuth(username, password)
	}

	// Get the data
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.Errorf("failed to fetch %s : %s", link, resp.Status)
	}

	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, resp.Body)
	body := buf.Bytes()

	return loadIndex(body)

}

// loadIndex loads an index file and does minimal validity checking.
//
// This will fail if API Version is not set (ErrNoAPIVersion) or if the unmarshal fails.
func loadIndex(data []byte) (*repo.IndexFile, error) {
	i := &repo.IndexFile{}
	if err := yaml.Unmarshal(data, i); err != nil {
		return i, err
	}

	i.SortEntries()
	if i.APIVersion == "" {
		return i, repo.ErrNoAPIVersion
	}
	return i, nil
}

// syncCharts create charts resource for a repo
// TODO: ace.ACE
func (r *ChartRepoReconciler) syncCharts(cr *v1beta1.ChartRepo, ctx context.Context) error {
	log := r.Log.WithValues("chartrepo", cr.GetName())

	checked := map[string]bool{}
	existCharts := map[string]v1beta1.Chart{}

	index, err := r.GetIndex(cr, ctx)
	if err != nil {
		return err
	}
	// this may causes bugs
	for name, _ := range index.Entries {
		checked[strings.ToLower(name)] = true
	}

	log.Info("retrieve charts from repo", "count", len(index.Entries))

	var charts v1beta1.ChartList
	labels := client.MatchingLabels{
		"repo": cr.GetName(),
	}
	if err := r.List(ctx, &charts, labels, client.InNamespace(r.Namespace)); err != nil {
		return err
	}
	for _, item := range charts.Items {
		name := strings.Split(item.GetName(), ".")[0]
		existCharts[name] = item
	}

	for on, versions := range index.Entries {
		name := strings.ToLower(on)
		chart := generateChartResource(versions, name, cr)
		// chart name can be uppercase in helm
		if _, ok := existCharts[name]; !ok {
			log.Info("chart not found, create", "name", cr.GetName()+"/"+on)
			err := r.Create(ctx, chart)
			if err != nil {
				if !apierrs.IsAlreadyExists(err) {
					return err
				}
			}
			continue
		}

		old := existCharts[name]
		if compareChart(old, chart) {
			chart.SetResourceVersion(old.GetResourceVersion())
			if err := r.Update(ctx, chart); err != nil {
				return err
			}
			log.Info("update chart", "name", old.Name, "repo", cr.Name)
		}

	}

	for name, item := range existCharts {
		if !checked[name] {
			dc := item
			dc.SetNamespace(cr.GetNamespace())
			if err := r.Delete(ctx, &dc); err != nil {
				return err
			}
			log.Info("delete charts", "name", item.GetName())
		}
	}

	return nil
}

// compareChart compare if a Chart need update
// 1. If length not equal, update
// 2. compare all digest

func compareChart(old v1beta1.Chart, new *v1beta1.Chart) bool {
	if len(old.Spec.Versions) != len(new.Spec.Versions) {
		return true
	}

	for _, o := range old.Spec.Versions {
		for _, n := range new.Spec.Versions {
			if o.Version == n.Version && o.Digest != n.Digest {
				return true
			}
		}
	}

	return false
}

func getChartName(repo, chart string) string {
	return fmt.Sprintf("%s.%s", strings.ToLower(chart), repo)
}

// generateChartResource create a Chart resource from the information in helm cache index
func generateChartResource(versions repo.ChartVersions, name string, cr *v1beta1.ChartRepo) *v1beta1.Chart {

	var vs []*v1beta1.ChartVersion
	for _, v := range versions {
		vs = append(vs, &v1beta1.ChartVersion{ChartVersion: *v})
	}

	spec := v1beta1.ChartSpec{
		Versions: vs,
	}

	labels := map[string]string{
		"repo": cr.GetName(),
	}
	if cr.GetLabels() != nil {
		project := cr.GetLabels()[util.ProjectKey]
		if project != "" {
			labels[util.ProjectKey] = project
		}
	}

	chart := v1beta1.Chart{
		TypeMeta: v1.TypeMeta{
			Kind:       "Chart",
			APIVersion: "app.alauda.io/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      getChartName(cr.GetName(), name),
			Namespace: cr.GetNamespace(),
			Labels:    labels,
			OwnerReferences: []v1.OwnerReference{
				*util.NewOwnerRef(cr, schema.GroupVersionKind{
					Group:   "app.alauda.io",
					Version: "v1alpha1",
					Kind:    "ChartRepo",
				}),
			},
		},
		Spec: spec,
	}

	return &chart

}

func (r *ChartRepoReconciler) updateChartRepoURL(ctx context.Context, cr *v1beta1.ChartRepo) error {
	old := cr.DeepCopy()
	mp := client.MergeFrom(old.DeepCopy())

	if old.Spec.URL == "" {
		old.Spec.URL = "http://captain-chartmuseum:8080/" + cr.GetName()
	}

	// save to origin object
	cr.Spec.URL = old.Spec.URL

	return r.Patch(ctx, old, mp)

}

// updateChartRepoStatus update ChartRepo's status
func (r *ChartRepoReconciler) updateChartRepoStatus(ctx context.Context, cr *v1beta1.ChartRepo, phase v1beta1.ChartRepoPhase, reason string) error {
	old := cr.DeepCopy()
	mp := client.MergeFrom(old.DeepCopy())

	old.Status.Phase = phase
	old.Status.Reason = reason

	if phase == v1beta1.ChartRepoSynced {
		now, _ := v1.Now().MarshalQueryParameter()
		if old.Annotations == nil {
			old.Annotations = make(map[string]string)
		}
		old.Annotations["alauda.io/last-sync-at"] = now

	}

	return r.Patch(ctx, old, mp)

}

func (r *ChartRepoReconciler) isReadyForResync(cr *v1beta1.ChartRepo) bool {
	log := r.Log.WithValues("chartrepo", cr.GetName())

	if cr.Status.Phase != "Synced" {
		return true
	}

	if cr.GetAnnotations() != nil && cr.GetAnnotations()["alauda.io/last-sync-at"] != "" {
		last := cr.GetAnnotations()["alauda.io/last-sync-at"]
		// see: https://stackoverflow.com/questions/25845172/parsing-date-string-in-go
		layout := "2006-01-02T15:04:05Z"
		t, err := time.Parse(layout, last)
		if err != nil {
			log.Error(err, "parse sync time error")
			return true
		}

		now := v1.Now()
		diff := now.Time.Sub(t).Seconds()

		//p, _ := now.MarshalQueryParameter()

		// log.Info("debug timer,", "last", last, "diff", diff)

		if diff >= 60 {
			return true
		}

		return false

	}
	return true

}
