import matplotlib.pyplot as plt

f1 = open("/home/sebas/Documentos/distribuidos/Distrubuted_computing/CUDA-Ver/2048 500 666 4.txt")

data = []
line = f1.readline()
while line != "":
    data.append(line.split())
    line = f1.readline()
print(data)
best = []
worst = []
mean = []
median = []
for i in range(len(data)-2):
    best.append(data[i+1][0])
    worst.append(data[i+1][1])
    mean.append(data[i+1][2])
    median.append(data[i+1][3])
best = [float(x) for x in best]
worst = [float(x) for x in worst]
mean = [float(x) for x in mean]
median = [float(x) for x in median]
plt.plot(best, label='best')
plt.plot(worst, label="worst")
plt.plot(mean, label="mean")
plt.plot(median, label="median")
plt.ylabel('fitness')
plt.xlabel('generations')
plt.legend( loc='upper right', borderaxespad=1)
plt.grid(True, linestyle='dotted', linewidth=1)
plt.show()
