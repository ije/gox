#!/bin/bash

read -p "please the deploy target('client' or 'server'): " target
if [ "$target" != "client" ] && [ "$target" != "server" ]; then
	echo "invalid the target ($target)..."
	exit
fi

read -p "please enter hostname or ip: " host
if [ "$host" == "" ]; then
	echo "miss the host..."
	exit
fi

hostSSHPort="22"
read -p "please enter vps host ssh port (default is '22'): " port
if [ "$port" != "" ]; then
	hostSSHPort="$port"
fi

supervisor="no"
read -p "install the supervisor config script('yes' or 'no', default is 'no')? " ok
if [ "$ok" == "yes" ]; then
	supervisor="yes"
fi

export GOOS=linux
export GOARCH=amd64

echo "--- compiling the x.tunnel.$target (linux_amd64)..."
go build x.tunnel.$target.go

echo "--- uploading..."
scp -P $hostSSHPort x.tunnel.$target root@$host:/usr/local/bin/_x.tunnel.$target
if [ "$supervisor" == "yes" ]; then
	scp -P $hostSSHPort x.tunnel.$target.supervisor.conf root@$host:/etc/supervisor/conf.d/x.tunnel.$target.conf
fi

echo "--- restart x.tunnel.$target..."
ssh -p $hostSSHPort root@$host << EOF
	supervisorctl stop x.tunnel.$target
	mv -f /usr/local/bin/_x.tunnel.$target /usr/local/bin/x.tunnel.$target
	chmod +x /usr/local/bin/x.tunnel.$target
	if [ "$supervisor" == "yes" ]; then
		supervisorctl reload
	else
		supervisorctl start x.tunnel.$target
	fi
EOF

rm x.tunnel.$target
