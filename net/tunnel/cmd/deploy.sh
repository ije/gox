#!/bin/bash

target="$1"
if [ "$target" != "client" ] && [ "$target" != "server" ]; then
	read -p "please enter the deploy target('client' or 'server'): " target
	if [ "$target" != "client" ] && [ "$target" != "server" ]; then
		echo "invalid the target '$target'..."
		exit
	fi
fi

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

initSupervisor="no"
read -p "setup supervisor program('yes' or 'no', default is 'no')? " ok
if [ "$ok" == "yes" ]; then
	initSupervisor="yes"
fi

exeArgs=""
supervisorConfDir="/etc/supervisor/conf.d"
if [ "$initSupervisor" == "yes" ]; then
	if [ "$target" == "server" ]; then
		read -p "please enter the server port(default is 333):" port
		if [ "$port" != "" ]; then
			exeArgs="$exeArgs -port=$port"
		fi
		read -p "please enter the server password: " password
		if [ "$password" != "" ]; then
			exeArgs="$exeArgs -password='$password'"
		fi
		read -p "please enter the server http port for status(default is 8080):" port2
		if [ "$port2" != "" ]; then
			exeArgs="$exeArgs -http-port=$port2"
		fi
	fi
	read -p "please enter the supervisor scripts directory(default is '$supervisorConfDir')? " dir
	if [ "$dir" != "" ]; then
		supervisorConfDir="$dir"
	fi
fi

sh build.sh $target
if [ "$?" != "0" ]; then 
	exit
fi

echo "--- uploading..."
scp -P $hostSSHPort $target-main/$target $loginUser@$host:/tmp/gox.tunnel.$target
if [ "$?" != "0" ]; then
	rm $target-main/$target
	exit
fi

scp -P $hostSSHPort install.sh $loginUser@$host:/tmp/tunnel.install.sh
if [ "$?" != "0" ]; then
	rm $target-main/$target
	exit
fi

echo "--- restart service..."
ssh -p $hostSSHPort $loginUser@$host << EOF
	echo "--- execute install script..."
	nohup sh /tmp/tunnel.install.sh $target "$exeArgs" $initSupervisor "$supervisorConfDir" >/dev/null 2>&1 &
EOF

rm $target-main/$target
