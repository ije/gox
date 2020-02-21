#!/bin/bash

supervisorctl status gox.tunnel.$1
if [ "$?" != "0" ]; then
	echo "needs supervisor"
	exit
fi

supervisorctl stop gox.tunnel.$1
rm -f /usr/local/bin/gox.tunnel.$1
mv -f /tmp/gox.tunnel.$1 /usr/local/bin/gox.tunnel.$1
chmod +x /usr/local/bin/gox.tunnel.$1
if [ "$2" = "yes" ]; then
	supervisorctl reload
else
	supervisorctl start gox.tunnel.$1
fi
