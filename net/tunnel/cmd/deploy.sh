#!/bin/bash

read -p "please enter the deploy target('client' or 'server'): " target
if [ "$target" != "client" ] && [ "$target" != "server" ]; then
	echo "invalid the target '$target'..."
	exit
fi

sh build.sh $target
if [ "$?" != "0" ]; then 
	exit
fi

read -p "please enter hostname or ip: " host
if [ "$host" = "" ]; then
	echo "missing the host..."
	rm x.tunnel.$target
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

initSupervisor="no"
read -p "install/update the supervisor config script('yes' or 'no', default is 'no')? " ok
if [ "$ok" = "yes" ]; then
	initSupervisor="yes"
fi

supervisordconfDir="/etc/supervisor"
if [ "$initSupervisor" = "yes" ]; then
	read -p "please enter the supervisord.conf directory(default is '$supervisordconfDir')? " dir
	if [ "$dir" != "" ]; then
		supervisordconfDir="$dir"
	fi
fi

echo "--- uploading..."
scp -P $hostSSHPort install.sh $loginUser@$host:/tmp/x.tunnel.install.sh
if [ "$?" != "0" ]; then
	rm x.tunnel.$target
	exit
fi

if [ "$initSupervisor" = "yes" ]; then
	scp -P $hostSSHPort x.tunnel.$target.supervisor.conf $loginUser@$host:$supervisordconfDir/conf.d/x.tunnel.$target.conf
	if [ "$?" != "0" ]; then
		rm x.tunnel.$target
		exit
	fi
fi

scp -P $hostSSHPort x.tunnel.$target $loginUser@$host:/tmp/x.tunnel.$target
if [ "$?" != "0" ]; then
	rm x.tunnel.$target
	exit
fi

echo "--- restart x.tunnel.$target..."
ssh -p $hostSSHPort $loginUser@$host << EOF
	echo "restart x.tunnel.$target ..."
	nohup sh /tmp/x.tunnel.install.sh $target $initSupervisor >/dev/null 2>&1 &
EOF

rm x.tunnel.$target
