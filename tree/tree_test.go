package tree

import (
	"fmt"
	"testing"
)

var keys_1	 []int    = []int {20,4,15,7,13,1,35};

var keys_2	 []int    = []int {20,19,18,17,16,15,14,13,12,11,10,9,8,7,6,5,4,3,2,1};

var keys_3	 []int    = []int {17,16,6,5,4,3,15,14,13,20,19,18,12,11,10,9,8,7,2,1};

var keys_4	 []int    = []int {1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,50,55,80,81,82,83,84,85,86,101,102,103,70,69,68,71,72,73,21,22,23,24,25,26};


func testCreateTestTree (nodeCount int, smallesKey int, largestKey int) *Tree {
//	var _tree  Tree = Tree {nil, nil};
	var _tree  Tree = Tree {nil};

	_keys := CreateRandomSlice (nodeCount, -smallesKey, largestKey);

	insertKeys (&_tree, _keys);
	return &_tree;
}

func insertKeys (tree *Tree, keys []int) {
	for _, _key := range keys {
		var _value string = fmt.Sprintf ("node value %d", _key);
		var newNode *Node = CreateNode (_key, &_value);
		tree.Insert (newNode);
		_err := testMerkle (tree);
		if (_err != nil) {
			fmt.Printf ("error in merkle after inserting key %d %s\n", _key, _err.Error ());
		}
	}
}


func TestTreeMeta_1 (t *testing.T) {
//	var _tree  Tree = Tree {nil, nil};
	var _tree  Tree = Tree {nil};
	insertKeys (&_tree, keys_1);
	testNodeMeta (t, _tree.Root);
	testNodeKeys (t, &_tree, keys_1);
}

func TestTreeMeta (t *testing.T) {
		var _tree  Tree = Tree {nil};
		insertKeys (&_tree, keys_4);
		testNodeMeta (t, _tree.Root);
		testNodeKeys (t, &_tree, keys_4);
		_err := testMerkle (&_tree);
		if (_err != nil) {
			t.Errorf ("%s\n", _err.Error ());
		}
	}
	
func TestDelete (t *testing.T) {
	var _tree  Tree = Tree {nil};
	insertKeys (&_tree, keys_4);
	_err := testMerkle (&_tree);
	if (_err == nil) {
		testNodeMeta (t, _tree.Root);
		testNodeKeys (t, &_tree, keys_4);
		_err = testMerkle (&_tree);
		_tree.Delete (50);
		testNodeMeta (t, _tree.Root);
		_err = testMerkle (&_tree);
		if (_err != nil) {
			t.Errorf ("%s\n", _err.Error ());
		}
	} else {
		t.Errorf ("error inserting keys. %s\n", _err.Error ());
	}
}
	
func testMerkle (tree *Tree) error {
	_, _err := tree.Root.checkMerkle ();
	return _err;
}

func testNodeKeys (t *testing.T, tree *Tree, keys []int) {
	for _, _key := range keys {
		if (_key == 103) {
			fmt.Printf ("looking for key %d\n", _key);
		}
        if (tree.Find (_key) == nil) {
			t.Errorf ("key %d not found\n", _key);
		}
    }
}


func testNodeMeta (t *testing.T, node *Node) int {
	var _errorText string = CheckNodeMeta(node);
	if (_errorText != "") {
		t.Errorf("following errors were detected %s\n", _errorText);
	}
	return 0;
}


