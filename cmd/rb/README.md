# RB

`rb` is a thin wrapper around the Rainbird API, making it suitable for terminal use.

## Usage

`rb help` gives a list of available commands and information on usage. An example of a complete session with the Speaks map might be:

```go
> export RB_SESSION=$(rb start aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee "") && echo $RB_SESSION
ffffffff-gggg-hhhh-iiii-jjjjjjjjjjjj
> rb query $RB_SESSION John speaks ""
QUESTION: Where does John live? (John - lives in - ?)
> rb response $RB_SESSION John 'lives in' England 100
ANSWERS:
  John - speaks - English [ 75%]
}
```
