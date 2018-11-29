import matplotlib.pyplot as plt


num_inserts = [2**i for i in range(6, 15)]
durations = [18420348, 39046089, 85578739, 185936039, 400919425, 836158053, 1987593546, 4107794378, 8029942148]
batched_durations = [0.01, 0.01, 0.02, 0.05, 0.11, 0.25, 500470515, 2156773276, 4631827869]

# Convert all durations to seconds
for i in range(len(durations)):
	durations[i] /= 10**9
	batched_durations[i] /= 10**9

plt.plot(num_inserts, durations, label="insert()",color="blue")
plt.plot(num_inserts, batched_durations, label="batchInsert()", color="red")
plt.legend()
plt.ylabel('Duration of random insert workload (seconds)')
plt.xlabel('Number of random inserts')
plt.xticks([i for i in num_inserts], num_inserts)
plt.title("Inserts v. Batched Inserts")
plt.savefig('go_smt.pdf', bbox_inches='tight')