import matplotlib.pyplot as plt


num_inserts = [2**i for i in range(6, 14)]
durations = [17207800, 35569878, 72824657, 150410290, 345796188, 763996854, 1840200469, 3721900874]

# Convert all durations to seconds
for i, el in enumerate(durations):
	durations[i] = el / 10**9

plt.plot(num_inserts, durations, "bo")
plt.ylabel('Duration of random insert workload (seconds)')
plt.xlabel('Number of random inserts')
plt.xticks([i for i in num_inserts], num_inserts)
plt.title("Durations of Single, Random Inserts")
plt.savefig('go_smt.pdf', bbox_inches='tight')