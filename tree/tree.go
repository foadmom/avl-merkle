package tree

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/rand"
	"sha_256"
	"strings"
	"time"
)

type Hash_Recalculation_Method  int;
const (
	NONE 			= iota+1
	SHALLOW
	DEEP
)
	

type Node struct {
    Key 		int;		
    Value		string;
	LastUpdated uint64;
    Right		*Node		`json:"-"`;  // ignore when marshalling
    Left		*Node		`json:"-"`;
    Height		int;
    Balance		int;
	Parent  	*Node		`json:"-"`;
	Hash		[]byte;		// `json:"-"`;
}

type Tree struct {
    Root		*Node;
}


var __debugTreePtr *Tree;

func CreateNode (key int, value *string) *Node {
	var _timeStamp uint64 = uint64 (time.Now ().UnixNano ());
    var _node  Node = Node{key, *value, _timeStamp, nil, nil, 0,0,nil,nil};

    return &_node;
}

func (tree *Tree) initialiseTree (rootNode *Node) {
	tree.Root = rootNode;
	tree.Root.nodeHash (DEEP);
	__debugTreePtr = tree;

}


func (tree *Tree) Insert (newNode *Node) error {
    if (tree.Root == nil) {
        // this must be the root node
        tree.initialiseTree (newNode);
    } else {
        tree.Root = tree.Root.insert(newNode);
        tree.Root.Balance = tree.Root.getBalance();
        tree.Root = tree.Root.reBalance();
    }
	tree.Root.checkMerkle();

    return nil;
}


func (node *Node) insert (newNode *Node) *Node {
    var _parent *Node = node.Parent;
    if newNode.Key == node.Key {
        // the key matches the parent key so
        // change the value of the parent node. ie update the parent 
		node.Value = newNode.Value;
		node.nodeHash (SHALLOW);
//		fmt.Printf ("duplicate key insert %d\n", node.Key);
        return node;
    } else if (newNode.Key > node.Key) {
        // the new node key is bigger than parent key
        // so add it to the right branch
        if node.Right == nil {
            node.Right = newNode;
            node.Right.Parent = node;
            if (node.Left == nil) {
                node.Height++;
            }
			node.nodeHash (DEEP);
        } else {
            node.Right = node.Right.insert(newNode);
            node.Height  = node.getHeight ();   // ??? not sure if this is correct
        }
    } else {
        // the newNode.key is less than the parent.key
        // so it must be added to the left branch
        if node.Left == nil {
            node.Left = newNode;
            node.Left.Parent = node;
            if (node.Right == nil) {
                node.Height++;
            }
			node.nodeHash (DEEP);
        } else {
            node.Left = node.Left.insert(newNode);
            node.Height  = node.getHeight ();   // ??? not sure if this is correct
        }
    }
    node.Balance = node.getBalance ();
    node = node.reBalance ();
	node.Parent = _parent;   // restore parent, just in case.
	node.nodeHash (SHALLOW);
	return node;
}


func (node *Node) reBalance () *Node {
    // save the current parent as this should not change
    // even though the node changes
    var _parent *Node = node.Parent; 
    if (node.Balance == 2) {
        if (node.Left != nil) {
            if (node.Left.Balance == 1) {
                node = node.rotateRight ();
            } else if (node.Left.Balance == -1) {
                // perform left and then right rotation
                node = node.rotateLeftRight ();
            }
            node.Parent = _parent;  // restore the parent
            node.Balance = node.getBalance ();
			node.nodeHash (DEEP);
        }
    }
    if (node.Balance == -2) {
        if (node.Right != nil) {
            if ( node.Right.Balance == -1) {
                node = node.rotateLeft ();
            } else if (node.Right.Balance == 1) {
                node = node.rotateRightLeft ();
            }
            node.Parent = _parent;  // restore the parent
            node.Balance = node.getBalance ();
			node.nodeHash (DEEP);
        }
    }
    return node;
}


// ==================================================================
// this is for the situation when nodes (height,balance)	3 (2,2) parent
// are arranged like this:							       /
// this is detected when parent node has                  2 (1,1) pivot
// a balance of 2, it's left child has a                 /
// balance of 1                                         1 (0,0)
//
//
//
// From;    3                to:  2
//         /                     / \
//        2                     1   3
//       / 
//      1 
// ==================================================================
func (parent *Node) rotateRight () *Node {
    var _pivot       *Node = parent.Left;
    var _pivotRight  *Node = _pivot.Right;

    // move parent to the right to become the pivot.Right and adjust 
    _pivot.Right = parent;

    parent.Height -= 2;  // parent has dropped 2 hops
    parent.Parent = _pivot;
    parent.Left = _pivotRight;
	if (_pivotRight != nil) {
		_pivotRight.Parent = parent;
	}

    parent.Balance = parent.getBalance();
    _pivot.Balance = _pivot.getBalance();

    return _pivot;
}

// ==================================================================
// opposite rotation to rotateLeft but similar logic.
// this occures when parent.Balance=-2 && parent.Right.Balance=-1
// From;    4                to:  8
//           \                   / \
//            8                 4  10
//             \                 
//              10               
// ==================================================================
func (parent *Node) rotateLeft () *Node{
    var _pivot      *Node = parent.Right;
    var _pivotLeft *Node = _pivot.Left;

    // move the parent to the left to become the pivot.Left and adjust
    _pivot.Left = parent;

	parent.Right = _pivotLeft;
	if (_pivotLeft != nil) {
		_pivotLeft.Parent = parent;
	}

    parent.Height -= 2;  // parent has dropped 2 hops
    parent.Parent = _pivot;

    _pivot.Balance = _pivot.getBalance();
    parent.Balance = parent.getBalance();
    return _pivot;
}

// ==================================================================
// this is equivalent of performing a left rotation followed by a right
// this occures when parent.Balance=2 && parent.Left.Balance=-1
//            A                       C
//           / \        ==>          / \ 
//          /   \                   /   \ 
//          B    Ar                /     \
//         / \                    B       A 
//        /   \                  / \     / \   
//        Bl   C                /   \   /   \    
//            / \              Bl   Cl  Cr   Ar
//           /   \                   
//           Cl  Cr                  
// ==================================================================
// Parent = A, _pivot = C
func (parent *Node) rotateLeftRight () *Node {
    var _pivot *Node = parent.Left.Right;   // pivot = C
            // pivot is what will be returned as parent.
            // so start setting it up from the top

    var _pivotRight = _pivot.Right;         // = old Cr
    var _pivotLeft  = _pivot.Left;          // = old Cl

    // start setting up _pivot from the top
    _pivot.Right = parent;        // pivot.Right = A
    _pivot.Right.Parent = _pivot;
    _pivot.Left  = parent.Left;   // pivot.Left  = B
    _pivot.Left.Parent = _pivot;
	_pivot.Height += 2;				// _pivote (C) has moved up 2 nodes

    // set up A
    _pivot.Right.Left    = _pivotRight  // A.Left becomes old Cr
    if (_pivotRight != nil) {
		_pivotRight.Parent = _pivot.Right
	}
    _pivot.Right.Height  = _pivot.Right.getHeight ();    
    _pivot.Right.Balance = _pivot.Right.getBalance ();

    // set up B
    _pivot.Left.Right  = _pivotLeft;     // B.Left = old Cl
    if (_pivotLeft != nil) {
		_pivotLeft.Parent = _pivot.Left
	}
    _pivot.Left.Height  = _pivot.Left.getHeight ();
    _pivot.Left.Balance = _pivot.Left.getBalance ();

    _pivot.Height   = _pivot.getHeight ();// adjust height. it's moving up
    _pivot.Balance  = _pivot.getBalance();

    return _pivot;
}

// ==================================================================
// this is equivalent of performing a right rotation followed by a left
// this occures when parent.Balance=-2 && parent.Right.Balance=1
//            A                            C                       
//           / \                          / \ 
//          /   \                        /   \ 
//         /     \           ==>        /     \
//        Al      B                    A       B 
//               /\                   / \     / \  
//              /  \                 /   \   /   \   
//             C   Br               Al   Cl  Cr   Br
//            / \            
//           /   \      
//          Cl  Cr      
// ==================================================================
func (parent *Node) rotateRightLeft () *Node {
    var _pivot *Node = parent.Right.Left;  // pivot = C
            // pivot is what will be returned as parent.
            // so start setting it up from the top

    var _pivotRight = _pivot.Right;         // = old Cr
    var _pivotLeft  = _pivot.Left;          // = old Cl
        
	// invalidate all the hashes in this node trio

    // start setting up _pivot from the top
    _pivot.Left = parent;        // pivot.Left = A
    _pivot.Left.Parent = _pivot;
    _pivot.Right  = parent.Right;   // pivot.Right  = B
	_pivot.Right.Parent = _pivot;
	_pivot.Height += 2;				// _pivote (C) has moved up 2 nodes

    // set up A
    _pivot.Left.Right    = _pivotLeft  // A.Left becomes old Cl
    if (_pivotLeft != nil) {
		_pivotLeft.Parent = _pivot.Left
	}
    _pivot.Left.Height  = _pivot.Left.getHeight ();    
	_pivot.Left.Balance = _pivot.Left.getBalance ();

    // set up B
    _pivot.Right.Left  = _pivotRight;     // B.Right = old Cr
    if (_pivotRight != nil) {
		_pivotRight.Parent = _pivot.Right
	}
    _pivot.Right.Height  = _pivot.Right.getHeight ();
    _pivot.Right.Balance = _pivot.Right.getBalance ();

    _pivot.Height   = _pivot.getHeight ();// adjust height. it's moving up
    _pivot.Balance  = _pivot.getBalance();

    return _pivot;
}

// ==================================================================
// ==================================================================
// ================         F I N D         =========================
// ==================================================================
// ==================================================================
// find a node with the key=?
// ==================================================================
func (tree *Tree) Find (key int) *Node {
	return tree.Root.find (key);
}


// ==================================================================
// find a node with the key=?
// ==================================================================
func (node *Node) find (key int) *Node {
	if (node.Key == key) {
		return node;
	} else if (key < node.Key) {
		if (node.Left != nil) {
			return (node.Left.find (key));
		} else {
			return nil;
		}
	} else {
		if (node.Right != nil) {
			return (node.Right.find (key));
		} else {
			return nil;
		}
	}
}


// ==================================================================
// find a the smallest value on the left side of this node
// ==================================================================
func (node *Node) FindSmallestOnRight () *Node {
	return node.Right.FindSmallestOnLeft();
}
// ==================================================================
// find a the smallest value on the left side of this node
// ==================================================================
func (node *Node) FindSmallestOnLeft () *Node {
	if (node.Left == nil) {
		return node;
	} else {
		return node.Left.FindSmallestOnLeft ();
	}
}

// ==================================================================
// find a the smallest value on the left side of this node
// ==================================================================
func (node *Node) FindLargestOnLeft () *Node {
	return node.Left.FindLargestOnRight();
}

// ==================================================================
// find a the smallest value on the left side of this node
// ==================================================================
func (node *Node) FindLargestOnRight () *Node {
	if (node.Right == nil) {
		return node;
	} else {
		return node.Right.FindLargestOnRight ();
	}
}

// ==================================================================
// ==================================================================
// ================       D E L E T E       =========================
// ==================================================================
func (tree *Tree) Delete (key int) {
	if (tree.Root != nil) {
		if (tree.Root.Key == key) {
			deleteRoot (tree);
		} else {
			tree.Root.Delete (key);
		}
	}
}

func deleteRoot (tree *Tree) {
	var _dummyValue string = "dummy Node, ignore";
	var _dummyParent *Node = CreateNode(0,&_dummyValue);

	_dummyParent.Key = tree.Root.Key - 1;
	_dummyParent.Left = tree.Root;
	// forcing the dummy parent to have a key larger than root
	// so the search goes the Left since it Left=Root and there
	// is no Right
	_dummyParent.Delete (tree.Root.Key);
}

// ==================================================================
// delete a single node in the tree. the structure under the node
// should not be deleted.
// ==================================================================
func (parent *Node) Delete (key int) *Node {
	var _deletedNode *Node;

	if (key < parent.Key) {
		// key must be in the left branch
		if (parent.Left != nil) {
			if (parent.Left.Key == key) {
				// we have found the node
				parent.delete (parent.Left);
			} else {
				_deletedNode = parent.Left.Delete (key);
			}
		}
	} else if (key > parent.Key) {
		// key must be in the left branch
		if (parent.Right != nil) {
			if (parent.Right.Key == key) {
				// we have found the node
				parent.delete (parent.Right);

			} else {
				_deletedNode = parent.Right.Delete (key);
			}
		}
	}

	parent.nodeHash (SHALLOW);
	parent.Height  = parent.getHeight();
	parent.Balance = parent.getBalance();

	return _deletedNode;
}

func (parent *Node) delete (node *Node)  *Node {

	var _deletedNode *Node;
	if (node.Left == nil && node.Right == nil) {
		return parent.delete_0 (node);
	} else if (node.Left != nil && node.Right != nil) {
		// more complex if the node to be deleted has both Left and Right child
		_deletedNode = parent.delete_4 (node);
	} else {
		// it has one child
		_deletedNode = parent.delete_2A3 (node);
	}
	parent.nodeHash (DEEP);
	return _deletedNode;
}


// ==================================================================
// delete case number 1 is when the node has no Left or Right children
// in the this example we want to delete Q who has one child, Q
//            A                            C                       
//           / \                          / \ 
//          /   \                        /   \ 
//         /     \           ==>        /     \
//        Al      M                    A       M 
//               /\                           / \  
//              /  \                         /   \   
//             /    \                       /     \   
//            J      P                     J       P
//           /        \                   /
//          /          \                 / 
//         J-l          Q              J-l
// ==================================================================
func (parent *Node) delete_0 (node *Node) *Node {
	// no children. can be deleted
	if (node == parent.Left) {
		parent.Left = nil;
	} else {
		parent.Right = nil;
	}
	return nil;
}


// ==================================================================
// delete case number 2 And 3 is when the node has 
// one and only one child. this is a little more complex than case 1
// in the this example we want to delete P who has one child, Q
//            A                            A                       
//           / \                          / \ 
//          /   \                        /   \ 
//         /     \           ==>        /     \
//        Al      M                    Al      M 
//               /\                           / \  
//              /  \                         /   \   
//             /    \                       /     \   
//            J      P                     J       Q
//           /        \                   /
//          /          \                 / 
//          Jl         Q                Jl
// ==================================================================
func (parent *Node) delete_2A3 (node *Node) *Node {
	var _replacementNode *Node;
	if (node.Left != nil) {
		_replacementNode = node.Left;
	} else {_replacementNode = node.Right;}

	if (node == parent.Left) {
		parent.Left = _replacementNode;
	} else {
		parent.Right = _replacementNode;
	}
	if (_replacementNode.Height > 0) {_replacementNode.Height--;}
	_replacementNode.Balance = _replacementNode.getBalance ();
	return nil;
}


// ==================================================================
// delete case number 4 is when the node has both Left and Right 
// children. this is a little more complex than case 2 & 3.
// in the this example we want to delete P who has one child, M
// M needs to be replaced, the replacement candidate is either
//   a. the smallest key on the Right of M or
//   b. the largest key on the Left of M
//
//        C                       C                       C            
//       / \                     / \                     / \ 
//      /   \                   /   \           OR      /   \ 
//     /     \         ==>     /     \          ==>    /     \
//    B       M               B       P               B       J 
//           / \                     / \                     / \  
//          /   \                   /   \                   /   \   
//         /     \                 /     \                 /     \  
//        J       P               J       Q               I       P
//       /         \             /                                 \
//      /           \           /                                   \
//     I             Q         I                                     Q
//
// ==================================================================
func (parent *Node) delete_4 (node *Node) *Node {
	var _replacementNode *Node;
	if (node.Right.Height > node.Left.Height) {
		_replacementNode = node.FindSmallestOnRight();
	} else {
		_replacementNode = node.FindLargestOnLeft();
	}
	node.Delete (_replacementNode.Key);
	node.Key   = _replacementNode.Key;
	node.Value = _replacementNode.Value;
	node.Height  = node.getHeight ();
	node.Balance = node.getBalance ();

	return nil;
}


// ==================================================================
// ==================================================================
// =================== U P D A T E ==================================
// ==================================================================
// ==================================================================
// ==================================================================
// update a node with new Value. There should not be any change to 
// tree structure.
// ==================================================================
func (tree *Tree) Update (key int, newValue string) {
	var _nodeToBeUpdated *Node = tree.Find (key);
	if (_nodeToBeUpdated != nil) {
		_nodeToBeUpdated.Value = newValue;
	}
}


// ==================================================================
// ==================================================================
// ==================================================================
// ==================================================================
// ==================================================================
// ==================================================================
// the height of a node is = Max (leftNode.Height, rightNode.Height)
// ==================================================================
func (node *Node) getHeight () int {
    var _leftHeight  int = -1;
    var _rightHeight int = -1;
    if (node.Left != nil) {
        _leftHeight = node.Left.Height;
    } 
    if (node.Right != nil) {
        _rightHeight = node.Right.Height;
    }
    return Max (_leftHeight, _rightHeight)+1;
}


// ==================================================================
// balance is caluculate as 'balance = left.Height - right.Height'
// ==================================================================
func (parent *Node) getBalance () int {
    var _leftHeight  int = 0;
    var _rightHeight int = 0;
    if (parent != nil) {
        if (parent.Left != nil) {
            _leftHeight = parent.Left.Height + 1;
        }
        if (parent.Right != nil) {
            _rightHeight = parent.Right.Height + 1;
        }
    }
    _balance := _leftHeight - _rightHeight;
    return _balance;
}

   
// ==================================================================
// 
// ==================================================================
func (node *Node) CalculateHeight () int {
	var _leftHeight  int;
	var _rightHeight int;
	var _nodeHeight  int;
	
	if (node.Left == nil) {
		_leftHeight = 0;
	} else {
		_leftHeight = node.Left.CalculateHeight ()+1;
	}
	if (node.Right == nil) {
		_rightHeight = 0;
	} else {
		_rightHeight = node.Right.CalculateHeight ()+1;
	}
	_nodeHeight = Max (_leftHeight, _rightHeight);

	return _nodeHeight;
}


// ==================================================================
// check all the metadata in the node 
// ==================================================================
func CheckNodeMeta (node *Node) string {
	var _leftHeight   int = -1;
	var _rightHeight  int = -1;
	var _errorText    string;

	if (node.Left != nil) { 
		_errorText = _errorText + CheckNodeMeta (node.Left);
	}
	if (node.Right != nil) {
		_errorText = _errorText + CheckNodeMeta (node.Right);
	}
	if (node.Left != nil) {
		_leftHeight = node.Left.CalculateHeight () ;
		if (_leftHeight != node.Left.Height) {
			_errorText = _errorText + fmt.Sprintf("error in the node Height, calculates = %d  node.Height = %d\n", _leftHeight, node.Left.Height);
		}
		if (node.Left.Parent != node) {
			_errorText = _errorText + fmt.Sprintf("error in the node Parent, calculates = %d  node.Left.Parent.key = %d\n", node.Key, node.Left.Parent.Key);
		}
	}
	if (node.Right != nil) {
		_rightHeight = node.Right.CalculateHeight () ;
		if (_rightHeight != node.Right.Height) {
			_errorText = _errorText + fmt.Sprintf("error in the node Height, calculates = %d  node.Height = %d\n", _rightHeight, node.Right.Height);
		}
		if (node.Right.Parent != node) {
			_errorText = _errorText + fmt.Sprintf("error in the node Parent, calculates = %d  node.Right.Parent.key = %d\n", node.Key, node.Right.Parent.Key);
		}
	}
	var _balance int = _leftHeight - _rightHeight;
	if (_balance != node.Balance) {
		_errorText = _errorText + fmt.Sprintf("error in the node Balance, calculates = %d  node.Balance = %d\n", _balance, node.Balance);
	}
	return _errorText
}
		

func (node *Node) CountNodes (count *int) int {
	*count++
	if (node.Left != nil) {
		node.Left.CountNodes (count);
	}
	if (node.Right != nil) {
		node.Right.CountNodes (count);
	}

	return *count;
}










// ==================================================================
// ==================================================================
// ==================================================================
// ==================================================================
// ==================================================================
// ==================================================================
// ==================================================================
// ==================================================================


// ==================================================================
// Code from https://appliedgo.net/balancedtree
// (c) 2016 Christoph Berger
// License: CC-BY-NC-SA, as this code is part of the blog article that is released under the same license.
// ==================================================================
// `Dump` dumps the structure of the subtree starting at node `n`, including node search values and balance factors.
// Parameter `i` sets the line indent. `lr` is a prefix denoting the left or the right child, respectively.
func (n *Node) Dump(i int, lr string) {
    if n == nil {
        return
    }
    indent := ""
    if i > 0 {
        //indent = strings.Repeat(" ", (i-1)*4) + "+" + strings.Repeat("-", 3)
        indent = strings.Repeat(" ", (i-1)*4) + "+" + lr + "--"
    }
    fmt.Printf("%s%d{%s}[%d, %d]\n", indent, n.Key, n.Value, n.Height, n.Balance)
    n.Left.Dump(i+1, "L")
    n.Right.Dump(i+1, "R")
}



// ==================================================================
// generate a random int between lower and upper
// ==================================================================
func generateRandomInt (lower int, upper int) int {
	rand.Seed (time.Now().UnixNano());
	var _rand int = rand.Intn(upper - lower) + lower;
	return _rand;
}

// ==================================================================
// generate a random int between lower and upper
// ==================================================================
func CreateRandomSlice (size int, lower int, upper int) []int {
	_slice := make([]int, size)
	var  	_index int;

	for _index < size  {
		_slice [_index] = generateRandomInt (lower, upper);
		_index++;
	}
	return _slice;
}


// ==================================================================
// ==================================================================
func Max(x, y int) int {
	if x > y {
	  return x
	}
	return y
	}
	
// ==================================================================
// ==================================================================
// ==================================================================
// ==================================================================
// ==================================================================
// ==================================================================
// ==================================================================
// return the hash for this node
// ==================================================================
func (node *Node) nodeHash (method Hash_Recalculation_Method) ([]byte, error) {
	var _nodeHash 		[]byte;
	var _err        error;
	var _leftHash   []byte;
	var _rightHash  []byte;

	if (node.Left != nil) {
		if (method == SHALLOW) {
			_leftHash = node.Left.Hash;
		} else {
			_leftHash, _err = node.Left.nodeHash (method);
		}
	}
	if (node.Right != nil && _err == nil) {
		if (method == SHALLOW) {
			_rightHash = node.Right.Hash;
		} else {
			_rightHash, _err = node.Right.nodeHash (method);
		}
	}

	if (_err == nil) {
		_nodeHash, _err = nodeHashCalc (node, _leftHash, _rightHash);
		if (_err == nil) {
			node.Hash = _nodeHash;
		} else {
			node.Hash = nil;
		}
	}
	return node.Hash, _err
}

// ==================================================================
// return the []byte that should be used for hashing
// ==================================================================
func whatNeedsToBeHashed_1 (node *Node, leftHash, rightHash []byte) []byte { 
	var _toBeHashed []byte;
	var _key 		[]byte = make ([]byte, 16);
	var _value 		[]byte;
	var _lastUpdated[]byte = make ([]byte, 16);;

	binary.LittleEndian.PutUint64(_key, uint64 (node.Key));
	binary.LittleEndian.PutUint64(_lastUpdated, node.LastUpdated);
	_value = []byte (node.Value);

	_toBeHashed = append (_key, _value...);
	_toBeHashed = append (_toBeHashed, _lastUpdated...);
	_toBeHashed = append (_toBeHashed, rightHash...);
	_toBeHashed = append (_toBeHashed, leftHash...);
	return (_toBeHashed);
}

// ==================================================================
// return the []byte that should be used for hashing
// for this to work node.Hash should be excluded from the json
// ==================================================================
func whatNeedsToBeHashed_2 (node *Node, leftHash, rightHash []byte) []byte { 
	var _toBeHashed []byte;
	var _nodeJson []byte;

	_nodeJson, _ = json.Marshal (node);
	_toBeHashed = append (_nodeJson, leftHash...);
	_toBeHashed = append (_toBeHashed, rightHash...);
	return (_toBeHashed);
}

// ==================================================================
// calculate the hash for this node
// ==================================================================
func nodeHashCalc (node *Node, leftHash, rightHash []byte) ([]byte, error) {
	var _err        error;
	var _hash       []byte;
	var _toBeHashed []byte;

	_toBeHashed = whatNeedsToBeHashed_1 (node, leftHash, rightHash);
	// _toBeHashed = whatNeedsToBeHashed_2 (node, leftHash, rightHash);
	_hash = []byte (sha_256.Sha_256 (_toBeHashed));

	return _hash, _err;

}

// ==================================================================
// calculates the hash for each node and checks it against the 
// existing hash to see if it is correct.
// ==================================================================
func (node *Node) checkMerkle () ([]byte, error) {
	var _rightMerkle []byte;
	var _leftMerkle  []byte;
	var _err 		 error;

	if (node.Left != nil) {
		_leftMerkle, _err = node.Left.checkMerkle();
	}
	if ((node.Right != nil) && (_err == nil)) {
		_rightMerkle, _err = node.Right.checkMerkle (); 
	}

	var _calculatedHash []byte;
	if  (_err == nil) {
		_calculatedHash, _err = nodeHashCalc (node, _leftMerkle, _rightMerkle);
		if (_err == nil) {
			if (bytes.Compare (_calculatedHash, node.Hash) != 0) {
				_err = fmt.Errorf ("\nERROR: node %d\n\tcalculated hash           % 2x\n does not match the node hash %x\n", node.Key, _calculatedHash, node.Hash);
			}
		}
	}

	return _calculatedHash, _err;
}

// ==================================================================
// copy attributes from A to node B
// ==================================================================
func copyNode (nodeA, nodeB *Node) *Node {
    nodeB.Key 		   = nodeA.Key;
    nodeB.Value		   = nodeA.Value;
	nodeB.LastUpdated  = nodeA.LastUpdated;
    nodeB.Right		   = nodeA.Right;
    nodeB.Left		   = nodeA.Left;
    nodeB.Height	   = nodeA.Height;
    nodeB.Balance	   = nodeA.Balance;
	nodeB.Parent  	   = nodeA.Parent;
//	nodeB.HashRecalcMethod	   = NONE;

	return nodeB;
}