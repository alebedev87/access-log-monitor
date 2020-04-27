#! /bin/bash

name="accesslog"

docker rm -f "${name}"
exec docker run --name "${name}" -d -p 8080:80 -v "${PWD}"/nginx:/var/log/nginx -v "${PWD}"/nginx.conf:/etc/nginx/nginx.conf:ro nginx
