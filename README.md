###

Это пока сборная солянка не готовая к промышленному использованию тк находится в глубокой R&D фазе MVP и поиску путей-идей-алгоритмов-решений.

### Debian deps:

```
apt install -y gcc golang make curl nginx
```


```
mkdir -p /usr/local/www
cp -a nginx /usr/local/www/nginx
```
```
cd send-lb
make
make install
```


```
cd send-data
./build.sh
```

```
cd /rest-graph-api
./build.sh
```

clone newbench (client side) to /root/newbench
cd /root/newbench/scripts/h2load_comparator
./build.sh
