#/bin/sh

TAG=`date +%Y%m%d%H%M%S`
PWD=`pwd`

echo $PWD

#go

GOPATH=$PWD

CMD="go get -d"

echo $CMD

$CMD

if [ $? -ne 0 ]; then
	echo -e "[FAIL] $CMD"
	exit
fi 

#build

CMD="docker pull registry.cn-hangzhou.aliyuncs.com/kk/kk-golang:latest"

echo $CMD

$CMD

if [ $? -ne 0 ]; then
	echo -e "[FAIL] $CMD"
	exit
fi

CMD="docker run --rm -v $PWD:/main:rw -v $PWD:/go:rw registry.cn-hangzhou.aliyuncs.com/kk/kk-golang:latest go build"

echo $CMD

$CMD

if [ $? -ne 0 ]; then
	echo -e "[FAIL] $CMD"
	exit
fi

#docker
CMD="docker build -t registry.cn-hangzhou.aliyuncs.com/kk/kk-httpd:$TAG ."

echo $CMD

$CMD

if [ $? -ne 0 ]; then
	echo -e "[FAIL] $CMD"
	exit
fi

CMD="docker push registry.cn-hangzhou.aliyuncs.com/kk/kk-httpd:$TAG"

echo $CMD

$CMD

if [ $? -ne 0 ]; then
	echo -e "[FAIL] $CMD"
	exit
fi


#cleanup

rm -rf src
rm -f main

