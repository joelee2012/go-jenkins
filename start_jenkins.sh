#!/bin/bash

REPO_ROOT=$(git rev-parse --show-toplevel)
echo "$REPO_ROOT"
function start {
  set -e
  local ip port password
  sudo docker run -dt --rm --name jenkins-master joelee123/standalone-jenkins:latest
  echo 'Waiting for Jenkins to start...'
  until sudo docker logs jenkins-master | grep -q 'Jenkins is fully up and running'; do
    sleep 1
  done
  ip=$(sudo docker inspect --format='{{.NetworkSettings.IPAddress}}' jenkins-master)
  password=$(sudo docker exec jenkins-master cat /var/jenkins_home/secrets/initialAdminPassword)
  version=$(sudo docker exec jenkins-master sh -c 'echo "$JENKINS_VERSION"')
  port=8080
  cat <<EOF | tee "${REPO_ROOT}"/env.sh
export JENKINS_URL='http://${ip}:${port}/'
export JENKINS_USER="admin"
export JENKINS_PSW='${password}'
export JENKINS_VERSION='${version}'
EOF
}

function stop {
  sudo docker stop jenkins-master
}

"$@"