from typing import List

class Transaction(object):
	def __init__(self, TransactionType: str, Index: str) -> None:
		self.MultiLeaf = False
		self.TransactionType = TransactionType
		self.Index = Index

	def __eq__(self, other):
		return self.Index == other.Index

	def __lt__(self, other):
		return self.Index < other.Index

	def from_dict(json_dict):
		tx = Transaction('', '')
		tx.__dict__ = json_dict
		return tx

class ReadTransaction(Transaction):
	def __init__(self, Index: str) -> None:
		Transaction.__init__(self, 'R', Index)

class WriteTransaction(Transaction):
	def __init__(self, Index: str, data: str) -> None:
		Transaction.__init__(self, 'W', Index)
		self.data = data

class MultiLeafTransaction(Transaction):
	def __init__(self, transactions: List[Transaction]) -> None:
		self.MultiLeaf = True
		self.transactions = transactions
		for Index in indices: 
			self.transaction_list.append(ReadTransaction.__init__(self, 'R', Index))

class MultiLeafReadTransaction(MultiLeafTransaction):
	def __init__(self, indices) -> None:
		transactions = []
		for Index in indices: 
			transactions.append(ReadTransaction(Index))
		MultiLeafTransaction.__init__(transactions)

class MultiLeafWriteTransaction(MultiLeafTransaction):
	def __init__(self, indices, datas) -> None:
		transactions = []
		for Index, data in zip(indices, data): 
			transactions.append(WriteTransaction(Index, data))
		MultiLeafTransaction.__init__(transactions)
