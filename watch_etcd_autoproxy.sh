#!/usr/bin/env bash

# generate initial configs
./autoproxy

# copy to correct path
# we store both versions for comparison later
sudo cp /etc/nginx/nginx.conf.tmp /etc/nginx/nginx.conf;

# and start nginx
sudo service nginx start

while :;
do
  # compare autoproxy-generated config to existing one
  is_diff=`diff --brief /etc/nginx/nginx.conf.tmp /etc/nginx/nginx.conf`

  if [ -z "$is_diff" ] ; then
    echo "no difference in files, not restarting nginx";
  else
    sudo cp /etc/nginx/nginx.conf.tmp /etc/nginx/nginx.conf;
    cat /etc/nginx/nginx.conf
    sudo service nginx reload;
  fi				
done
