make clean
make build

name=results128-1000
touch $name

for size in 100 150 439 666 1002
do
    echo $size >> $name
    for th in 1 2 4 8 16
    do
        for ((i=0;i<5;i++))
        do
            ./main $th setup/$size.tsp >> $name
        done
        echo \ >> $name
    done
done




