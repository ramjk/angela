from typing import List

class Transaction(object):
	def __init__(self, transaction_type: str, index: str) -> None:
		self.multi_leaf = False
		self.transaction_type = transaction_type
		self.index = index

	def __eq__(self, other):
		return self.index == other.index

	def __lt__(self, other):
		return self.index < other.index


class ReadTransaction(Transaction):
	def __init__(self, index: str) -> None:
		Transaction.__init__(self, 'R', index)

class WriteTransaction(Transaction):
	def __init__(self, index: str, data: str) -> None:
		Transaction.__init__(self, 'W', index)
		self.data = data

class MultiLeafTransaction(Transaction):
	def __init__(self, transactions: List[Transaction]) -> None:
		self.multi_leaf = True
		self.transactions = transactions
		for index in indices: 
			self.transaction_list.append(ReadTransaction.__init__(self, 'R', index))

class MultiLeafReadTransaction(MultiLeafTransaction):
	def __init__(self, indices) -> None:
		transactions = []
		for index in indices: 
			transactions.append(ReadTransaction(index))
		MultiLeafTransaction.__init__(transactions)

class MultiLeafWriteTransaction(MultiLeafTransaction):
	def __init__(self, indices, datas) -> None:
		transactions = []
		for index, data in zip(indices, data): 
			transactions.append(WriteTransaction(index, data))
		MultiLeafTransaction.__init__(transactions)
