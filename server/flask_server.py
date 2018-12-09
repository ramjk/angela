import argparse
import ray
import random
import timeit

from flask import Flask, jsonify, request
from werkzeug.exceptions import BadRequest
from math import log

from server.smt_api import getLatestRootDigest, read, batch_insert
from common.util import random_index, random_string
# from server.transaction import Transaction, WriteTransaction
from server.worker import Worker
# from common.util import random_index, random_string, send_data

app = Flask(__name__)

num_workers = None
epoch_length = None
tree_depth = None

ray_info = dict()
request_count = 0
epoch_number = 1

def calculate_worker_id(index: str, prefix_length: int) -> int:
	return int(index, 2) >> (len(index) - prefix_length)

@app.route("/merkletree/bench", methods=['GET'])
def benchmark():
	global request_count
	num_inserts = int(request.args.get('num_inserts'))

	print("Generating workload...")
	indices = [random_index() for i in range(num_inserts)]
	data = [random_string() for i in range(num_inserts)]

	for i in range(num_inserts):
		worker_id = calculate_worker_id(indices[i], ray_info['prefix_length'])
		print("Sending insert request to worker {}".format(worker_id))
		ray.get(ray_info['leaf_workers'][worker_id].queue_write.remote(indices[i], data[i]))

	request_count += num_inserts
	print("Workload generated")
	return "Workload generated", 200

@app.route("/merkletree/root", methods=['GET'])
def get_root():
	worker_id = random.randint(0, num_workers)
	# 9-th "worker" is the root
	if worker_id < num_workers-1:
		worker_response = ray.get(ray_info['leaf_workers'][worker_id].process_read.remote(""))
	else:
		worker_response = ray.get(ray_info['root_worker'].process_read.remote(""))

	print("[get_root]", worker_response)
	return worker_response

@app.route("/merkletree/prove", methods=['GET'])
def generate_proof():
	multi_leaf = request.args.get('multi_leaf')
	transaction_type = request.args.get('transaction_type')
	index = request.args.get('index')

	if not (multi_leaf and transaction_type and index):
		raise BadRequest("For proof, specify multi_leaf, transaction_type, index")

	worker_id = random.randint(0, num_workers)
	if worker_id < num_workers-1:
		worker_response = ray.get(ray_info['leaf_workers'][worker_id].process_read.remote(index))
	else:
		worker_response = ray.get(ray_info['root_worker'].process_read.remote(index))

	return jsonify(worker_response.__dict__)

@app.route("/merkletree/update", methods=['POST'])
def update_leaf():
	global request_count, epoch_number

	multi_leaf = request.form.get('multi_leaf')
	transaction_type = request.form['transaction_type']
	index = request.form['index']
	data = request.form['data']

	if not (multi_leaf and transaction_type and index):
		raise BadRequest("For proof, specify multi_leaf, transaction_type, index")
	
	# Ray insert
	worker_id = calculate_worker_id(index, ray_info['prefix_length'])
	try:
		# Queue up write request in child. If ray get works, then its commited.
		object_id = ray_info['leaf_workers'][worker_id].queue_write.remote(index, data)
		ray.get(object_id)
		request_count += 1
	except Exception as err:
		print(err)
		raise BadRequest("Uhh something went wrong...")

	if request_count % epoch_length == 0:
		object_ids = list()
		worker_roots = list()

		for leaf_worker in ray_info['leaf_workers']:
			object_ids.append(leaf_worker.batch_update.remote(epoch_number))

		for object_id in object_ids:
			worker_root_digest, prefix = ray.get(object_id)
			worker_roots.append((prefix, worker_root_digest))

		output = ray.get(ray_info['root_worker'].batch_update.remote(epoch_number, worker_roots))
		epoch_number += 1 # FIXME: This should be a persistent value

	print("Epoch number:", epoch_number)
	return "Request for inserting has been committed", 200

if __name__ == '__main__':
	parser = argparse.ArgumentParser()
	parser.add_argument('num_workers', type=int, help="number of worker nodes")
	parser.add_argument('epoch_length', type=int, help="number of transactions per epoch")
	parser.add_argument('tree_depth', type=int, help="depth of tree")
	parser.add_argument('epoch_number', type=int)
	args = parser.parse_args()

	num_workers = args.num_workers
	epoch_length = args.epoch_length
	tree_depth = args.tree_depth
	epoch_number = args.epoch_number

	ray.init()#redis_address="localhost:6379")

	prefix_length = int(log(num_workers-1, 2))
	root_worker = Worker.remote(prefix_length-1, -1, prefix_length)
	leaf_workers = list()

	for worker_id in range(num_workers-1):
		worker = Worker.remote(tree_depth-prefix_length, worker_id, prefix_length, root_worker)
		leaf_workers.append(worker)

	ray_info['leaf_workers'] = leaf_workers

	root_worker.set_children.remote(leaf_workers)
	ray_info['root_worker'] = root_worker
	ray_info['prefix_length'] = prefix_length

	app.run()#host='0.0.0.0')