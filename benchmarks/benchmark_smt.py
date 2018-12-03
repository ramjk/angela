import matplotlib.pyplot as plt


num_inserts = [2**i for i in range(6, 15)]
durations = [18204421, 40471915, 92862276, 186704565, 419569214, 1018730535, 2234560254, 4107110821, 9293492478]
batched_durations = [0.01, 0.03, 0.03, 0.06, 0.22, 0.26, 1107994840, 1972901594, 4623583503]
batched_2_durations = [0.01, 0.01, 0.02, 0.05, 0.23, 0.26, 1019380233, 2074988222, 4787574386]

# Convert all durations to seconds
for i in range(len(durations)):
	durations[i] /= 10**9
	batched_durations[i] /= 10**9
	batched_2_durations[i] /= 10**9

plt.plot(num_inserts, durations, label="insert()",color="blue")
plt.plot(num_inserts, batched_durations, label="batchInsert()", color="red")
plt.plot(num_inserts, batched_2_durations, label="batch2Insert()", color="green")
plt.legend()
plt.ylabel('Duration of random insert workload (seconds)')
plt.xlabel('Number of random inserts')
plt.xticks([i for i in num_inserts], num_inserts)
plt.title("Inserts v. Batched Inserts")
plt.gcf().set_size_inches(24, 12, forward=True)
plt.savefig('go_smt.pdf', bbox_inches='tight')