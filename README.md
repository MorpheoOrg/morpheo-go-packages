Morpheo-Compute: a container-oriented Machine-Learning job runner
=================================================================

This repository holds the Golang code for the core of the morpheo project. It
contains the code for our storage API (which is simply a frontend API for a blob
storage such as a hard drive or Amazon S3, targeted at storing problem and
algorithms as containers and data files as... files :)

TL;DR
-----
* `client`: Golang API client for `storage` and a `fabric hyperledger peer`.  Important note: The fabric-sdk-go client is required to use this package, consequently the docker image running your go builds need the following libraries intalled: libtool libltdl-dev.
* `common`: data structure definitions and common interfaces and types
  (container runtime backend, blob store backend, broker backend...). Code in
  this folder should not import any other library in the Morpheo project.
* `utils/dind-daemon` defines an alpine based docker image running the Docker
  daemon. The compute workers run containers (problem workflow & algo) in this
  "Docker in Docker" container.

License
-------

All this code is open source and licensed under the CeCILL license - which is an
exact transcription of the GNU GPL license that also is compatible with french
intellectual property law. Please find the attached licence in English [here](./LICENSE) or
[in French](./LICENCE).

Note that this license explicitely forbids redistributing this code (or any
fork) under another licence.

Maintainers
-----------
* Max-Pol Le Brun <maxpol_a t_morpheo.co>
* Étienne Lafarge <etienne_a t_rythm.co>
