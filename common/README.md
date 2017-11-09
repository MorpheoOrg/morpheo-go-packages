Dreemcare Common Golang Libraries
=================================

This repository contains Golang code common to all the Golang services of the
Morpheo platform.

 * **Blobstore**: blob storage abstraction (and its local disk and S3
   implementations)
 * **Broker**: broker abstration (and its NSQ implementation)
 * **Container Runtime**: container runtime abstraction (and its `docker`
   implementation).

In addition, a `MultiStringFlag` type has been defined, all the data
structures necessary for the project are defined in this folder
(`data_structures.go`).

**Note**: if possible all code in this repo should have no internal dependency
on other morpheo libs.

Maintainers
-----------
 * Étienne Lafarge <etienne_a t_rythm.co>
 * Max-Pol Le Brun <max-pol_a t_morpheo.io>
