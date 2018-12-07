# Knuts

Knative utilities for install and management (i.e. one-time setup and operator
concerns, rather than ongoing development tools).

## Usage

Right now, build templates are the only install supported. At some point, there
will be an `install` command for the base components, and an `events` command to
set up eventing.

```
$ knuts builds
? Which build templates do you want to install  [Use arrows to move, type to filter]
  [ ]  bazel: Bazel with container_push rule
  [ ]  buildah: Buildah mechanism for building from Dockerfiles. Requires $BUILDER_IMAGE set in your Build.
  [x]  buildpack: Buildpack
> [x]  jib-gradle: Gradle build with JIB
  [ ]  jib-maven: Maven build with JIB
  [ ]  kaniko: Dockerfile with Kaniko
Dry run: `kubectl --filename "https://raw.githubusercontent.com/knative/build-templates/master/buildpack/buildpack.yaml"`
Dry run: `kubectl --filename "https://raw.githubusercontent.com/knative/build-templates/master/jib/jib-gradle.yaml"`
          
? Which registries to push to  [Use arrows to move, type to filter]
  [ ]  docker: Docker (user secret)
  > [x]  gcr.io: Google Container Registry
? GCP Project to push images to _myproject_
...
```

**By default, `knuts` runs in a "dry run" mode where it won't make any
  changes. Use the `--dry_run=false` flag to apply the changes to your
  cluster.**

## WARNING

This is totally work in progress, and right now does nothing. I'll be adding
commands and utilities in (possibly) the following order:

1. Install [build templates](https://github.com/knative/build-templates) - DONE
2. Create [docker](https://github.com/knative/build-templates/tree/master/gcr_helper) & github secrets for build templates - DONE for GCR, TODO for docker & others
3. [Install Knative](https://github.com/knative/docs/tree/master/install)
4. Help find & install [eventing Sources](https://github.com/knative/eventing-sources)
5. Upgrade
