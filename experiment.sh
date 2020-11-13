for popSize in 128 256 512 1024 2048
do
    for gen in 100 200 500 1000
    do
        sed -i "s/const popSize = [0-9]*/const popSize = $popSize/" main.go
        make clean
        make build
        
        name=results/$popSize-$gen
        touch $name

        for size in 100 150 439 666 1002
        do
            echo $size >> $name
            for th in 1 2 4 8 16 32
            do
                for ((i=0;i<5;i++))
                do
                    ./main $th $gen setup/$size.tsp $i >> $name
                done
                echo \ >> $name
            done
        done
    done
done
