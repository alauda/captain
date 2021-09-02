# ARM64 Support
Captain has built in docker image support for arm64 platform, even though it's not automatically generated from GitHub CI.

Replace the images in `deploy.yaml` before apply it:

```bash
# since v1.2.1
alaudapublic/captain:<version>-arm64 
# cert-init image is removed since v1.3.0
alaudapublic/captain-cert-init:v1.0-arm64
alaudapublic/chartmuseum:v2.0-arm64
```