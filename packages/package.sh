VERSION=0.1.3_dev
# -X sets the value of a string variable in main, others are size optimizations
go build -ldflags "-w -s -O -X main.version $VERSION" ../
# use FPM to build a package with the correct version number
~/.gem/ruby/2.2.0/bin/fpm -s dir -t rpm -n distributive -f -d go -m langston@aster.is -v $VERSION --epoch 0 ../distributive=/usr/bin/ ../samples=/etc/distributive.d/samples/
