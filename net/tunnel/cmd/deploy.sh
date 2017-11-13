#!/bin/bash

read -p "please enter the deploy target('client' or 'server'): " target
if [ "$target" != "client" ] && [ "$target" != "server" ]; then
	echo "invalid the target '$target'..."
	exit
fi

goos="linux"
read -p "please enter the deploy operating system(default is 'linux'): " sys
if [ "$sys" != "" ]; then
	goos="$sys"
fi
export GOOS=$goos

goarch="amd64"
read -p "please enter the deploy OS Architecture(default is 'amd64'): " arch
if [ "$arch" != "" ]; then
	goarch="$arch"
fi
export GOARCH=$goarch

read -p "please enter hostname or ip: " host
if [ "$host" == "" ]; then
	echo "missing the host..."
	exit
fi

loginUser="root"
read -p "please enter the host ssh login user (default is 'root'): " user
if [ "$user" != "" ]; then
	loginUser="$user"
fi

hostSSHPort="22"
read -p "please enter the host ssh port (default is '22'): " port
if [ "$port" != "" ]; then
	hostSSHPort="$port"
fi

supervisor="no"
read -p "install/update the supervisor config script('yes' or 'no', default is 'no')? " ok
if [ "$ok" == "yes" ]; then
	supervisor="yes"
fi

echo "--- compiling the x.tunnel.$target (${goos}_$goarch)..."
go build x.tunnel.$target.go
if [ "$EXCODE" != "0" ]; then 
	exit
fi
# exit

echo "--- uploading..."
scp -P $hostSSHPort x.tunnel.$target $loginUser@$host:/usr/local/bin/_x.tunnel.$target
if [ "$supervisor" == "yes" ]; then
	scp -P $hostSSHPort x.tunnel.$target.supervisor.conf $loginUser@$host:/etc/supervisor/conf.d/x.tunnel.$target.conf
fi

echo "--- restart x.tunnel.$target..."
ssh -p $hostSSHPort $loginUser@$host << EOF
	supervisorctl stop x-tunnel-$target
	mv -f /usr/local/bin/_x.tunnel.$target /usr/local/bin/x.tunnel.$target
	chmod +x /usr/local/bin/x.tunnel.$target
	if [ "$supervisor" == "yes" ]; then
		supervisorctl reload
	else
		supervisorctl start x-tunnel-$target
	fi
EOF

rm x.tunnel.$target
