supervisorctl stop x-tunnel-$1
if [ "$?" != "0" ]; then
	echo "need supervisor"
	exit
fi
mv -f /tmp/x.tunnel.$1 /usr/local/bin/x.tunnel.$1
chmod +x /usr/local/bin/x.tunnel.$1
if [ "$2" = "yes" ]; then
	supervisorctl reload
else
	supervisorctl start x-tunnel-$1
fi
