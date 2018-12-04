from typing import List

class Transaction(object):
	def __init__(self, TransactionType: str, index: str) -> None:
		self.MultiLeaf = False
		self.TransactionType = TransactionType
		self.Index = index

	def __eq__(self, other):
		return self.Index == other.Index

	def __lt__(self, other):
		return self.Index < other.Index

	def from_dict(json_dict):
		tx = Transaction('', '')
		tx.__dict__ = json_dict
		return tx

class ReadTransaction(Transaction):
	def __init__(self, index: str) -> None:
		Transaction.__init__(self, 'R', index)

class WriteTransaction(Transaction):
	def __init__(self, index: str, data: str) -> None:
		Transaction.__init__(self, 'W', index)
		self.Data = data

class MultiLeafTransaction(Transaction):
	def __init__(self, transactions: List[Transaction]) -> None:
		self.MultiLeaf = True
		self.transactions = transactions
		for index in indices: 
			self.transaction_list.append(ReadTransaction.__init__(self, 'R', Index))

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
