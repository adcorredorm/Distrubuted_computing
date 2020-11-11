build:
	nvcc --ptxas-options=-v --compiler-options '-fPIC' -o libmain.so --shared main.cu
	sudo mv libmain.so /usr/local/lib
	sudo ldconfig
	go build main.go

clean:
	rm main