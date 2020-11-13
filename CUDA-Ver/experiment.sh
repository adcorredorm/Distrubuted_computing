make clean
make build

name=results439
touch $name

for population in 128 256 512 1024 2048
do
    echo $population >> $name
    for th in 100 200 500 1000
    do
        for i in 1 2 3 4 5
        do
            ./main $population $th ../setup/439.tsp $i >> $name
        done
        echo \ >> $name
    done
done




