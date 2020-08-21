# geomag

geomagnetic collection programs

* __slgeomag__ SeedLink raw csv collector
* __msgeomag__ MiniSeed raw csv collector
* __wsgeomag__ FDSN raw csv collector

To compile C library dependencies, first run:

```
make -C internal/mseed clean all
make -C internal/slink clean all
```
