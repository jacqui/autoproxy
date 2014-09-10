#!/usr/bin/env bash

# generate initial configs
./autoproxy &

# copy to correct path
# we store both versions for comparison later
sudo cp /etc/nginx/nginx.conf.tmp /etc/nginx/nginx.conf;

# and start nginx
sudo service nginx start

while :;
do
  # compare autoproxy-generated config to existing one
  if diff --brief /etc/nginx/nginx.conf.tmp /etc/nginx/nginx.conf; then
    sleep 10;
  else
    echo "copying /etc/nginx/nginx.conf.tmp -> /etc/nginx/nginx.conf";
    sudo cp /etc/nginx/nginx.conf.tmp /etc/nginx/nginx.conf;
    cat /etc/nginx/nginx.conf
    echo "reloading nginx...";
    sudo service nginx reload;
    echo "done!";
  fi				
done
