
geard-router-haproxy 
====================

This repo can be built and used as a plugin to the geard routing system (https://github.com/openshift/geard). 'geard' creates the intermediate routing structure that is picked up by this package when run as a docker container. Three steps to use this router -

	1. Build - on a system that has go installed and GOPATH set -
		go get github.com/openshift/geard
		go get github.com/openshift/geard-router-haproxy
		cd $GOPATH/src/github.com/openshift/geard-router-haproxy
		./build
	2. Prepare the docker image using the Dockerfile. This will pull in the latest haproxy source code and compile it.
	3. Run a container that uses the above image file (see run.example).

Change something?
	Modify the default_pub_keys.pem, or the haproxy_template.conf and run the ./build script again. No need to redo the Docker image.


