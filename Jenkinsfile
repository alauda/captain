// https://jenkins.io/doc/book/pipeline/syntax/
@Library('alauda-cicd') _

// global variables for pipeline
// image can be used for promoting...
def IMAGE
def RELEASE_VERSION
def RELEASE_BUILD
def release
pipeline {
	agent { label 'golang-1.12' }

	options {
		buildDiscarder(logRotator(numToKeepStr: '10'))
		disableConcurrentBuilds()
		skipDefaultCheckout()
	}

	parameters {
		booleanParam(name: 'DEBUG', defaultValue: false, description: 'Debug will not do final changes...')
	}
	//(optional) 环境变量
	environment {
		FOLDER = 'src/alauda.io/captain'

		// for building an scanning
		REPOSITORY = "captain"
		OWNER = "mathildetech"

		BITBUCKET_FEEDBACK_ACCOUNT = "alaudabot"
		SONARQUBE_BITBUCKET_CREDENTIALS = "alaudabot"
		IMAGE_REPOSITORY = "index.alauda.cn/alaudaorg/captain"
		IMAGE_CREDENTIALS = "alaudacn-daniel"
		DINGDING_BOT = "devops-chat-bot"
		TAG_CREDENTIALS = "alaudabot-bitbucket"
		PROXY_CREDENTIALS_ID = 'proxy'

		// go lang 1.12 proxy and modules support
        GO111MODULE = "auto"
        // GOPROXY = "https://athens.acp.alauda.cn"
		GOPATH = "${WORKSPACE}"

		// charts pipeline name
		CHARTS_PIPELINE = "/common/common-charts-pipeline"
		CHART_NAME = "captain"
		CHART_COMPONENT = "captain"
	}
	// stages
	stages {
		stage('Checkout') {
			steps {
				script {
					dir(FOLDER) {
						container('tools') {
							// checkout code
							withCredentials([
								usernamePassword(credentialsId: PROXY_CREDENTIALS_ID, passwordVariable: 'PROXY_ADDRESS', usernameVariable: 'PROXY_ADDRESS_PASS')
							]) { PROXY_CREDENTIALS = "${PROXY_ADDRESS}" }
							sh "git config --global http.proxy ${PROXY_CREDENTIALS}"
							def scmVars
							retry(2) { scmVars = checkout scm }
							release = deploy.release(scmVars)

							RELEASE_BUILD = release.version
							RELEASE_VERSION = release.majorVersion
							// echo "release ${RELEASE_VERSION} - release build ${RELEASE_BUILD}"
							echo """
								release ${RELEASE_VERSION}
								version ${release.version}
								is_release ${release.is_release}
								is_build ${release.is_build}
								is_master ${release.is_master}
								deploy_env ${release.environment}
								auto_test ${release.auto_test}
								environment ${release.environment}
								majorVersion ${release.majorVersion}
							"""
							// copying kubectl from tools
							sh "cp /usr/local/bin/kubectl ."
						}
					}
				}
			}
		}
		stage('CI'){
			failFast true
			parallel {
				stage('Build') {
					steps {
						script {
							dir(FOLDER) {
								container('golang') {
                                    sh 'make build'
                                }
								container('tools') {
								    sh "upx captain"

									IMAGE = deploy.dockerBuild(
										"Dockerfile", //Dockerfile
										".", // build context
										IMAGE_REPOSITORY, // repo address
										RELEASE_BUILD, // tag
										IMAGE_CREDENTIALS, // credentials for pushing
									)
									// start and push
									IMAGE.start().push()
								}
							}
						}
					}
				}
				stage('Unit test') {
					steps {
						script {
							dir(FOLDER) {
								// running unit tests and gathering coverage
								container('golang') {
                                    sh "make test"
								}
							}
						}
					}
				}
			}
		}
		stage('Tag git') {
			when {
				expression {
					release.shouldTag()
				}
			}
			steps {
				script {
					dir(FOLDER) {
						container('tools') {
							deploy.gitTag(
								TAG_CREDENTIALS,
								RELEASE_BUILD,
								OWNER,
								REPOSITORY
							)
						}
					}
				}
			}
		}
		stage('Chart Update') {
			when {
				expression {
                    // TODO: Change when charts are ready
					release.shouldUpdateChart()
				}
			}
			steps {
				script {
					echo "will trigger charts-pipeline using branch ${release.chartBranch}"

					build job: CHARTS_PIPELINE, parameters: [
						[$class: 'StringParameterValue', name: 'CHART', value: CHART_NAME],
						[$class: 'StringParameterValue', name: 'VERSION', value: RELEASE_VERSION],
						[$class: 'StringParameterValue', name: 'COMPONENT', value: CHART_COMPONENT],
						[$class: 'StringParameterValue', name: 'IMAGE_TAG', value: RELEASE_BUILD],
						[$class: 'BooleanParameterValue', name: 'DEBUG', value: false],
						[$class: 'StringParameterValue', name: 'ENV', value: release.environment],
						[$class: 'StringParameterValue', name: 'BRANCH', value: release.chartBranch],
						[$class: 'StringParameterValue', name: 'PR_CHANGE_BRANCH', value: release.change["branch"]],
					], wait: true
				}
			}
		}
	}

	post {
		success {
			dir(FOLDER) {
				script {
					container('tools') {
						def msg = "流水线完成了"
						deploy.notificationSuccess(REPOSITORY, DINGDING_BOT, msg, RELEASE_BUILD)
						if (release != null) { release.cleanEnv() }
					}
				}
			}
		}

		failure {
			dir(FOLDER) {
				script {
					container('tools') {
						deploy.notificationFailed(REPOSITORY, DINGDING_BOT, "流水线失败了", RELEASE_BUILD)
						if (release != null) { release.cleanEnv() }
					}
				}
			}
		}
	}
}



