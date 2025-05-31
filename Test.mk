HOST = localhost:8000
DATA_STORE = .store/images/
EXAMPLE_IMAGES = debian alpine ubuntu fedora archlinux busybox nginx redis mysql postgres python node httpd golang ruby hello-world

clean:
	sudo rm -rf .store debian_latest.sif script.sh

clean-jobs:
	sudo rm -rf .store/jobs

debian_latest.sif:
	apptainer pull docker://debian:latest

image-put: debian_latest.sif
	curl -X PUT -i http://$(HOST)/images/debian/raw -F "file=@debian_latest.sif"

image-get:
	curl -X GET -i http://$(HOST)/images/debian/raw -o debian_latest.sif

image-properties:
	curl -X GET -i http://$(HOST)/images/debian/properties

image-get-redirect:
	curl -X GET -i http://$(HOST)/images/debian

image-delete:
	curl -X DELETE -i http://$(HOST)/images/debian

image-put-all:
	for image in $(EXAMPLE_IMAGES); do \
		apptainer pull .store/images/$$image.sif docker://$$image; \
	done

image-list:
	curl -X GET -i http://$(HOST)/images/

image-list-page-2:
	curl -X GET http://$(HOST)/images/?skip=10 | jq .

image-list-format:
	curl -X GET http://$(HOST)/images/ | jq .

job-post:
	curl -X POST -i http://$(HOST)/jobs/

JOB_ID = 1

job-put-properties:
	curl -H 'Content-type: application/json' -X PUT -i http://$(HOST)/jobs/$(JOB_ID)/properties -d '{"image": "debian", "artifacts": ["test", "test2"]}'

job-get:
	curl -X GET -i http://$(HOST)/jobs/$(JOB_ID)/

script.sh:
	echo "#!/bin/bash" > script.sh
	echo "apt update -qqy" >> script.sh
	echo "apt install -qqy neofetch" >> script.sh
	echo "neofetch > test" >> script.sh
	echo "echo hello > test2" >> script.sh
	echo "neofetch" >> script.sh

job-put-script: script.sh
	curl -X PUT -i http://$(HOST)/jobs/$(JOB_ID)/script/ -F "file=@script.sh"

job-get-script:
	curl -X GET -i http://$(HOST)/jobs/$(JOB_ID)/script/

job-get-state:
	curl -X GET -i http://$(HOST)/jobs/$(JOB_ID)/state/

job-put-state: 
	curl -X PUT -i "http://$(HOST)/jobs/1/state/?state=start"

job-get-log:
	curl -X GET -i http://$(HOST)/jobs/$(JOB_ID)/log/

job-get-artifacts:
	curl -X GET -i http://$(HOST)/jobs/$(JOB_ID)/artifacts/

job-get-artifacts-raw:
	curl -X GET -i  --output artifacts.zip http://$(HOST)/jobs/$(JOB_ID)/artifacts/data

jobs:
	curl -X GET -I http://$(HOST)/jobs/
	curl -X GET http://$(HOST)/jobs/ | jq .

jobs-filter:
	curl -X GET http://$(HOST)/jobs/?state=done | jq .

pipeline:
	curl -X POST -i http://$(HOST)/pipelines/ -H 'Content-type: application/json' -d '@pipeline.json'
