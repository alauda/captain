package release

var releaseCRDYaml = `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: releases.app.alauda.io
spec:
  group: app.alauda.io
  names:
    kind: Release
    listKind: ReleaseList
    plural: releases
    singular: release
    shortNames:
      - rel
  scope: Namespaced
  versions:
  - name: v1alpha1
    additionalPrinterColumns:
    - name: Status
      type: string
      jsonPath: .status.status
    - name: Age
      type: date
      jsonPath: .metadata.creationTimestamp
    schema:
      openAPIV3Schema:
        properties:
          apiVersion:
            description: 'APIVersion defines the versioned schema of this representation
              of an object. Servers should convert recognized schemas to the latest
              internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
            type: string
          kind:
            description: 'Kind is a string value representing the REST resource this
              object represents. Servers may infer this from the endpoint the client
              submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
            type: string
          metadata:
            type: object
          spec:
            description: ReleaseSpec describes a deployment of a chart, together with
              the chart and the variables used to deploy that chart.
            properties:
              chartData:
                description: ChartData is the chart that was released.
                type: string
              configData:
                description: ConfigData is the set of extra Values added to the chart.
                  These values override the default values inside of the chart.
                type: string
              hooksData:
                description: Hooks are all of the hooks declared for this release.
                type: string
              manifestData:
                description: ManifestData is the string representation of the rendered
                  template.
                type: string
              name:
                type: string
              version:
                description: Version is an int which represents the version of the
                  release.
                type: integer
            type: object
          status:
            properties:
              Description:
                description: Description is human-friendly "log entry" about this
                  release.
                type: string
              deleted:
                description: Deleted tracks when this object was deleted.
                type: string
              first_deployed:
                description: Deleted tracks when this object was deleted.
                type: string
              last_deployed:
                description: LastDeployed is when the release was last deployed.
                type: string
              notes:
                description: Contains the rendered templates/NOTES.txt if available
                type: string
              status:
                description: Status is the current state of the release
                type: string
            type: object
        required:
        - spec
        type: object
    served: true
    storage: true
`
