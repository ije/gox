#!/bin/bash

supervisorctl status tunnel.$1
if [ "$?" != "0" ]; then
	echo "needs supervisor"
	exit
fi

supervisorctl stop tunnel.$1
rm -f /usr/local/bin/tunnel.$1
mv -f /tmp/tunnel.$1 /usr/local/bin/tunnel.$1
chmod +x /usr/local/bin/tunnel.$1
if [ "$2" = "yes" ]; then
	supervisorctl reload
else
	supervisorctl start tunnel.$1
fi
