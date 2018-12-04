import matplotlib.pyplot as plt
import merkle.smt
import timeit
import cProfile, pstats
from common import util

setup = '''
from common import util
from server.server import Server
from client.client import Client
server = Server(8088, 9, 1000, 256)
client = Client('localhost', 8088)

for i in range(999):
	client.insert_leaf(util.random_index(), util.random_string())
	
'''
stmt = "client.insert_leaf(util.random_index(), util.random_string())"

duration = timeit.timeit(stmt=stmt, setup=setup)
print(duration)