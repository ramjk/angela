* Add mock tests to TestSparseMerkleTree so that we can isolate verification from generating the path.
	* We can use a smaller, hard-coded copath and use it as an argument in verify_copath.
	* Probably don't actually need to use mocking because we can just set the root_digest ourselves.