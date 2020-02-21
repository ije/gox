#!/bin/bash

target="$1"
if [ "$target" != "client" ] && [ "$target" != "server" ]; then
	read -p "please enter the deploy target('client' or 'server'): " target
	if [ "$target" != "client" ] && [ "$target" != "server" ]; then
		echo "invalid the target '$target'..."
		exit
	fi
fi

sh build.sh $target
if [ "$?" != "0" ]; then 
	exit
fi

read -p "please enter hostname or ip: " host
if [ "$host" = "" ]; then
	echo "missing the host..."
	rm tunnel.$target
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

supervisordConfDir="/etc/supervisor/conf.d"
if [ "$initSupervisor" = "yes" ]; then
	read -p "please enter the supervisord.conf directory(default is '$supervisordConfDir')? " dir
	if [ "$dir" != "" ]; then
		supervisordConfDir="$dir"
	fi
fi

echo "--- uploading..."
scp -P $hostSSHPort install.sh $loginUser@$host:/tmp/tunnel.install.sh
if [ "$?" != "0" ]; then
	rm tunnel.$target
	exit
fi

if [ "$initSupervisor" = "yes" ]; then
	scp -P $hostSSHPort $target/supervisor.conf $loginUser@$host:$supervisordConfDir/gox.tunnel.$target.conf
	if [ "$?" != "0" ]; then
		rm $target/$target
		exit
	fi
fi

scp -P $hostSSHPort $target/$target $loginUser@$host:/tmp/gox.tunnel.$target
if [ "$?" != "0" ]; then
	rm $target/$target
	exit
fi

echo "--- restart service..."
ssh -p $hostSSHPort $loginUser@$host << EOF
	echo "tunnel $target restarted"
	nohup sh /tmp/tunnel.install.sh $target $initSupervisor >/dev/null 2>&1 &
EOF

rm $target/$target
