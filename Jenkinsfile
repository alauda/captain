library "alauda-cicd"
def language = "golang"
AlaudaPipeline {
	config = [
		agent: 'golang-1.13',
		folder: '.',
		chart: [
            [
                chart: "captain",
                pipeline: "chart-captain",
                project: "acp",
                component: "captain",
            ],
		],
		scm: [
			credentials: 'acp-acp-gitlab'
		],
		docker: [
			repository: "acp/captain",
			credentials: "alaudak8s",
			context: ".",
			dockerfile: "Dockerfile",
		],
		sonar: [
		    binding: "sonarqube",
	            enabled: true,
		],
		notification: [
	    	name: "default"
		],

	]
	env = [
		GO111MODULE: "on",
		GOPROXY: "https://athens.alauda.cn",
	]
	yaml = "alauda.yaml"
	stepsYaml =
		"""
      steps:
      - name: "Unit test"
        container: "golang"
        groovy:
        - |+
          try {
            sh script: "make test", label: "unit tests..."
          } finally {
            junit allowEmptyResults: true, testResults: 'pkg/**/*.xml'
          }
      - name: "Build"
        container: "golang"
        commands:
        - |+
          make build
      """
}
