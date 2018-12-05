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
setup = '''
from common import util
from server.transaction import WriteTransaction
from server.server import Server
from client.client import Client
server = Server(8088, 9, 1000, 256)
client = Client('localhost', 8088)

for i in range(999):
	transaction = WriteTransaction(util.random_index(), util.random_string())
	server.receive_transaction(transaction)
'''
stmt = '''
transaction = WriteTransaction(util.random_index(), util.random_string())
server.receive_transaction(transaction)
server.close()
'''

duration = timeit.timeit(stmt=stmt, setup=setup, number=1)
print(duration)