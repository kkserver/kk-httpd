#/bin/sh

TAG=`date +%Y%m%d%H%M%S`
PWD=`pwd`


exitCommand() {
	rm -rf src
	rm -f main
	exit
}

runCommand() {
	echo $CMD
	$CMD
	if [ $? -ne 0 ]; then
		echo -e "[FAIL] $CMD"
		exitCommand
	fi 
}

buildProject() {

	GOPATH=$PWD

	CMD="go get -d"

	runCommand

	#build

	CMD="docker pull registry.cn-hangzhou.aliyuncs.com/kk/kk-golang:latest"

	runCommand

	CMD="docker run --rm -v $PWD:/main:rw -v $PWD:/go:rw registry.cn-hangzhou.aliyuncs.com/kk/kk-golang:latest go build"

	runCommand

	#docker
	CMD="docker build -t $PROJECT:$TAG ."
	runCommand

	CMD="docker push $PROJECT:$TAG"
	runCommand

	CMD="docker tag $PROJECT:$TAG $PROJECT:latest"
	runCommand

	CMD="docker push $PROJECT:latest"
	runCommand

}

echo $PWD

#go

PROJECT="registry.cn-hangzhou.aliyuncs.com/kk/kk-httpd"
buildProject

#exit

echo "[OK] TAG: $TAG"

exitCommand

