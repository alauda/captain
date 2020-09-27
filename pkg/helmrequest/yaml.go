package helmrequest

var helmRequestCRDYaml = `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: helmrequests.app.alauda.io
spec:
  group: app.alauda.io
  version: v1alpha1
  names:
    kind: HelmRequest
    listKind: HelmRequestList
    plural: helmrequests
    singular: helmrequest
    shortNames:
      - hr
  additionalPrinterColumns:
    - name: Chart
      type: string
      description: The chart of this HelmRequest
      JSONPath: .spec.chart
    - name: Version
      type: string
      description: Version of this chart
      JSONPath: .spec.version
    - name: Namespace
      type: string
      description: The namespace which the chart deployed to
      JSONPath: .spec.namespace
    - name: AllCluster
      type: boolean
      description: Is this chart will be installed to all cluster
      JSONPath: .spec.installToAllClusters
    - name: Phase
      type: string
      description: The phase of this HelmRequest
      JSONPath: .status.phase
    - name: Age
      type: date
      JSONPath: .metadata.creationTimestamp
  scope: Namespaced
  subresources:
    status: {}
  validation:
    # openAPIV3Schema is the schema for validating custom objects.
    openAPIV3Schema:
      properties:
        spec:
          description: HelmRequestSpec defines the deploy info of a helm chart
          type: object
          required:
            - chart
          properties:
            chart:
              type: string
              description: Chart is a helm chart name ,in the format of <repo>/<chart>
            namespace:
              type: string
              description: Namespace is the namespace this chart will be installed to. If not set, consider it's metadata.namespace
            releaseName:
              type: string
              description: ReleaseName is the Release name. If not set, consider it's metadata.name
            clusterName:
              type: string
              description: ClusterName is the target cluster name, where this chart will be installed to. If not set, this chart will be installed to the current cluster.
            dependencies:
              type: array
              description: Dependencies defines the HelmRequest list this HelmRequest will depends to, it will wait for them to be Synced
              items:
                type: string
            installToAllClusters:
              description: InstallToAllClusters decide if we want to install this chart to all cluster.
              type: boolean
            values:
              type: object
              nullable: true
              description: Values defines custom values for this chart
            version:
              type: string
              description: Version defines the chart version
            valuesFrom:
              type: array
              description: ValuesFrom defines the config file we want to ref to. In kubernetes, this will be ConfigMap/Secret
              items:
                type: object
                properties:
                  configMapKeyRef:
                    type: object
                    required:
                      - name
                    description: ConfigMapKeyRef defines a ref to a ConfigMap(in the same namespace)
                    properties:
                      name:
                        type: string
                      key:
                        type: string
                      optional:
                        type: boolean
                  secretKeyRef:
                    type: object
                    required:
                      - name
                    description: SecretKeyRef defines a ref to a Secret(in the same namespace)
                    properties:
                      name:
                        type: string
                      key:
                        type: string
                      optional:
                        type: boolean
`
