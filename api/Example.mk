URL:=localhost:4000
EXAMPLE_IMAGE:=test

get-images:
	curl ${URL}/images -H "Accept: application/json" | jq .

post-images:
	curl -L -i -X POST ${URL}/images \
		-H "Content-Type: application/json" \
		-d '{"name": "${EXAMPLE_IMAGE}", "description": "An example image"}'

get-image:
	curl -i ${URL}/images/${EXAMPLE_IMAGE} -H "Accept: application/json"

delete-image:
	curl -L -i -X DELETE ${URL}/images/${EXAMPLE_IMAGE}

get-image-properties:
	curl -i ${URL}/images/${EXAMPLE_IMAGE}/properties -H "Accept: application/json"

patch-image-properties:
	curl -L -i -X PATCH ${URL}/images/${EXAMPLE_IMAGE}/properties \
		-H "Content-Type: application/json" \
		-d '{"description": "An updated example image"}'

put-image-raw:
	curl -L -i -X PUT ${URL}/images/${EXAMPLE_IMAGE}/raw \
	-H 'Content-Type: multipart/form-data' \
	--compressed -F "data=@/home/mateusz/Images/qemu/debian/clean-ssh.qcow2.zst"
