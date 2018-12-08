import matplotlib.pyplot as plt
import merkle.smt
import timeit
import cProfile, pstats
from common import util

# setup = '''
# from common import util
# from server.server import Server
# from client.client import Client
# server = Server(8088, 9, 1000, 256)
# client = Client('localhost', 8088)

# for i in range(999):
# 	client.insert_leaf(util.random_index(), util.random_string())
	
# '''

num_inserts = 256

setup = '''
import time
from client.client import Client
from common.util import random_index, random_string

Client.benchmark({})
time.sleep(2)
index = random_index()
data = random_string() 
'''
durations=[]
for i in range(5):
	print("Iteration {}".format(i))
	res = timeit.timeit(stmt="Client.insert_leaf(index, data)", setup=setup.format(num_inserts-1), number=1)
	print("Duration is {}".format(res))
	durations.append(res)
print("Durations", durations)
print("Average time", sum(durations)/len(durations))

# durations_8 = [1.765327169999999, 2.6414496519999986, 4.583681233, 8.643308655999999, 15.488765152999996, 32.219870587]
# num_inserts_8 = [64, 128, 256, 512, 1024, 2048]

# durations_16 = [2.220073276, 2.8402085510000017, 4.688701079000001, 8.511718733000002, 16.375632815000003, 31.240891469999994]
# num_inserts_16 = [64, 128, 256, 1024, 2048]

# plt.plot(num_inserts, durations, label="w/ Ray (Local)",color="blue")
# plt.xticks([i for i in num_inserts], num_inserts)
# plt.title("Durations of Batched, Random Inserts")
# plt.ylabel('Duration of random insert workload (seconds)')
# plt.xlabel('Number of random inserts')
# plt.gcf().set_size_inches(24, 12, forward=True)
# plt.savefig('benchmarks/ray.pdf', bbox_inches='tight')