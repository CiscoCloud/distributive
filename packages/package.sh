go build -ldflags "-w -s -O" ../ &&
~/.gem/ruby/2.2.0/bin/fpm -s dir -t rpm -n distributive -f -d go -m langston@aster.is -v 0.1.2 --epoch 0 ../distributive=/usr/bin/ ../samples=/etc/distributive.d/samples/
