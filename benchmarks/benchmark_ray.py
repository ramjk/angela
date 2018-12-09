import matplotlib.pyplot as plt
import merkle.smt
import timeit
import cProfile, pstats
import subprocess
import time
import os
import ray

from math import log
from common import util
from client.client import Client

max_batch = 1024
batches = [2**i for i in range(6, int(log(max_batch, 2)) + 1)]
global_soundness_error = 0
epoch_number = 1

setup = '''
import time
from client.client import Client
from common.util import random_index, random_string

Client.benchmark({})
time.sleep(2)
index = random_index()
data = random_string() 
'''

servers = dict()
for i in range(3, 6):
	print("[server w/ {} Workers]".format(2**i))
	servers[i] = batch_size_durations = dict()
	for batch_size in batches:
		print("    [batch_size {}]".format(batch_size))
		
		pid = subprocess.Popen(['python', '-m', 'server.flask_server', str(2**i + 1), str(batch_size), str(256), str(epoch_number)])#, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
		time.sleep(10)
		durations = list()

		for j in range(5):
			print("        [Iteration {}]".format(j))
			res = timeit.timeit(stmt="Client.insert_leaf(index, data)", setup=setup.format(batch_size-1), number=1)
			durations.append(res)

		avg_dur = sum(durations)/len(durations)
		print("    [batch_size {}] Durations: {}".format(batch_size, durations))
		print("    [batch_size {}] Average time: {} seconds".format(batch_size, sum(durations)/len(durations)))
		batch_size_durations[batch_size] = avg_dur
		pid.kill()
		ray.shutdown()
	print("[server w/ {} Workers]".format(2**i))

# durations_8 = [1.765327169999999, 2.6414496519999986, 4.583681233, 8.643308655999999, 15.488765152999996, 32.219870587]
# num_inserts_8 = [64, 128, 256, 512, 1024, 2048]

# durations_16 = [2.220073276, 2.8402085510000017, 4.688701079000001, 8.511718733000002, 16.375632815000003, 31.240891469999994]
# num_inserts_16 = [64, 128, 256, 1024, 2048]

# colors = ["blue", "green", "red"]

for i, batch_size_durations in servers.items():
	plt.plot(batches, batch_size_durations.values(), label="Ray distr. over {} EC2 instances".format(2**i + 1), marker="o")

plt.rcParams.update({'font.size': 22})
plt.legend(prop={'size': 20})
plt.xticks([i for i in batches], batches)
plt.title("Durations of Batched, Random Inserts")
plt.ylabel('Duration of batched insert (seconds)')
plt.xlabel('Batch size')
plt.gcf().set_size_inches(24, 12, forward=True)
plt.savefig('benchmarks/ray.pdf', bbox_inches='tight')