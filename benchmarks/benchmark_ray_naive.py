import matplotlib.pyplot as plt
import merkle.smt
import timeit
import cProfile, pstats
from common import util

setup = '''
from common import util
from server.transaction import WriteTransaction
from server.server import Server
from client.client import Client
server = Server(8088, 9, 1, 256, {})
client = Client('localhost', 8088)

for i in range({}):
	print("insert")
	transaction = WriteTransaction(util.random_index(), util.random_string())
	server.receive_transaction(transaction)
'''
stmt = '''
transaction = WriteTransaction(util.random_index(), util.random_string())
server.receive_transaction(transaction)
'''
durations = list()
num_inserts = list()
flag = True
for i in range(6, 12):
	durations.append(timeit.timeit(stmt=stmt, setup=setup.format(flag, 2**i - 1), number=1))
	flag = False
	num_inserts.append(2**i)
plt.plot(num_inserts, durations, label="w/ Ray (Local)",color="blue")
plt.xticks([i for i in num_inserts], num_inserts)
plt.title("Durations of Single, Random Inserts")
plt.ylabel('Duration of random insert workload (seconds)')
plt.xlabel('Number of random inserts')
plt.gcf().set_size_inches(24, 12, forward=True)
plt.savefig('benchmarks/ray_naive.pdf', bbox_inches='tight')