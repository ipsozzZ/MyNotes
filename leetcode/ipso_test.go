package main

import (
	"fmt"
	"math"
	"math/big"
	"sort"
	"testing"
	"time"
)

func TestHello(t *testing.T) {
	t.Log("TestHello")
	var flag = true
	if flag {
		t.Log("11111")
	}else {
		t.Log("2222222")
	}
}


func TestChan(t *testing.T) {

	go say("world")
	say("hello")

	fmt.Println("---------------1")

	a := []int{7, 2, 8, -9, 4, 0}

	c := make(chan int)
	go sum(a[:len(a)/2], c)
	go sum(a[len(a)/2:], c)
	x, y := <-c, <-c // receive from c

	fmt.Println(x, y, x+y)

	fmt.Println("---------------2")

	c2 := make(chan int, 2)
	c2 <- 1
	c2 <- 2
	fmt.Println(<-c2)
	fmt.Println(<-c2)

	fmt.Println("---------------3")
	c3 := make(chan int, 10)
	go fibonacci(cap(c3), c3)
	for i := range c3 {
		fmt.Println(i)
	}

	fmt.Println("---------------4")
	c4 := make(chan int)
	quit := make(chan int)
	go func() {
		for i := 0; i < 10; i++ {
			fmt.Println(<-c4)
		}
		quit <- 0
	}()
	fibonacci2(c4, quit)

	fmt.Println("---------------5")
	tick := time.Tick(100 * time.Millisecond)
	boom := time.After(500 * time.Millisecond)
	for {
		select {
		case <-tick:
			fmt.Println("tick. ")
		case <-boom:
			fmt.Println("BOOM!")
			return
		default:
			fmt.Println("    .")
			time.Sleep(50 * time.Millisecond)
		}
	}

}

func say(s string) {
	for i := 0; i < 5; i++ {
		time.Sleep(100 * time.Millisecond)
		fmt.Println(s)
	}
}

func sum(a []int, c chan int) {
	sum := 0
	for _, v := range a {
		sum += v
	}
	c <- sum // send sum to c
}

func fibonacci(n int, c chan int) {
	x, y := 0, 1
	for i := 0; i < n; i++ {
		c <- x
		x, y = y, x+y
	}
	close(c)
}

func fibonacci2(c, quit chan int) {
	x, y := 0, 1
	for {
		select {
		case c <- x:
			x, y = y, x+y
		case <-quit:
			fmt.Println("quit")
			return
		}
	}
}




/************************************************ leetcode ************************************************/


type TreeNode struct {
     Val int
     Left *TreeNode
     Right *TreeNode
}

/**
leetcode 100
 */

func TestLC100(t *testing.T) {
	tree1 := &TreeNode{Val: 1, Left: &TreeNode{Val: 2}, Right: &TreeNode{Val: 3, Right: &TreeNode{Val: 4}}}
	tree2 := &TreeNode{Val: 1, Left: &TreeNode{Val: 2}, Right: &TreeNode{Val: 3, Left: &TreeNode{Val: 4}}}
	var sameTree = isSameTree(tree1,tree2)
	if sameTree {
		t.Log("true")
	}else {
		t.Log("false")
	}
}

func isSameTree(p *TreeNode, q *TreeNode) bool {
	if p == nil && q == nil {
		return true
	}
	if p == nil || q == nil {
		return false
	}
	if p.Val != q.Val {
		return false
	}
	return isSameTree(q.Left, p.Left) && isSameTree(q.Right, p.Right)
}

// leetcode 101. 对称二叉树
func TestLC101(t *testing.T) {
	tree := &TreeNode{Val: 0, Left: nil, Right: &TreeNode{Val: 1, Left: nil, Right: nil}}
	t.Log(isSymmetric(tree))
}

// isSymme
func isSymmetric(root *TreeNode) bool {
	if root == nil {
		return true
	}
	return isSymme(root.Left, root.Right)
}

func isSymme(L *TreeNode, R *TreeNode) bool {
	if L == nil && R == nil {
		return true
	}
	if L == nil || R == nil {
		return false
	}
	if L.Val == R.Val {
		return isSymme(L.Left, R.Right) && isSymme(L.Right, R.Left)
	}
	return false
}


func TestMyTt1(t *testing.T) {
	str := "gqm"

	f := 1
	for i,v := range MyFunc(f) {
		t.Logf("i %d, v== %v \n", i, v)
	}

	t.Logf("iii == %s \n", str)

	t.Logf("%v", time.Now().UnixNano())
}



func MyFunc(flag int) []*int {

	if flag == 1 {
		a := 1
		b := 2
		c := 3
		return []*int{&a,&b,&c}
	}else {
		return nil
	}
}

// 双指针、滑动窗口
func removeDuplicates(nums []int) int {
	nlen := len(nums)
	if nlen < 2 {
		return nlen
	}
	pre := nums[0]
	j := 1
	for i := 1; i < nlen; i++ {
		if nums[i] != pre {
			nums[j] = nums[i]
			pre = nums[j]
			j++
		}
	}
	return j
}

func findLengthOfLCIS(nums []int) int {
	nlen := len(nums)
	res, j := 0,0
	for i := 0; i < nlen; i++ {
		if i > 0 && nums[i-1] >= nums[i] {
			j = i
		}
		res = isMax(res, i-j)
	}
	return res
}

func isMax(a, b int) int {
	if a >= b {
		return a
	}
	return b
}

// 704 二分查找
func search(nums []int, target int) int {
	left := 0
	right := len(nums) -1
	for left <= right {
		mid := (left + right) /2
		if nums[mid] > target {
			right = mid - 1
		} else if nums[mid] < target {
			left = mid + 1
		}else {
			return mid
		}
	}
	return -1
}

func Test704(t *testing.T) {
	args := []int{-1,0,1,3,5,9,12}
	t.Logf("search === %d", search(args, -3))
}


// 278 第一个错误的版本
//func firstBadVersion(n int) int {
//	begin := 1
//	end := n
//
//	for begin <= end {
//		mid := (begin + end) / 2
//		if isBadVersion(mid) == true {
//			if mid-1 >= 1 && isBadVersion(mid-1) == false {
//				return mid
//			}
//			if mid == 1 {
//				return mid
//			}
//			end = mid - 1
//		}else if isBadVersion(mid) == false {
//			if mid+1 <= n && isBadVersion(mid+1) == true {
//				return mid+1
//			}
//			if mid == n {
//				return mid
//			}
//			begin = mid + 1
//		}
//	}
//	return n
//}

func Test278(t *testing.T) {
	// isBadVersion ??????????
}


// 35 搜索插入位置
func searchInsert(nums []int, target int) int {
	begin := 0
	end := len(nums) - 1

	for begin <= end {
		mid := (end + begin) / 2
		if nums[mid] > target {
			if begin == mid {
				return begin
			}
			 end = mid - 1
		}else if nums[mid] < target {
			if end == mid {
				return end + 1
			}
			begin = mid + 1
		}else {
			return mid
		}
	}
	return end
}

func searchInsert1(nums []int, target int) int {
	n := len(nums)
	left, right := 0, n - 1
	ans := n
	for left <= right {
		mid := (right - left) >> 1 + left
		if target <= nums[mid] {
			ans = mid
			right = mid - 1
		} else {
			left = mid + 1
		}
	}
	return ans
}

func Test35(t *testing.T) {
	args := []int{-1,0,1,3,5,9,12}
	//t.Logf("search === %d", searchInsert1(args, 13))
	var res []int
	res = append(res, args[3+1:]...)
	t.Logf("search === %+v", args[:3])
	t.Logf("search === %+v", args)
	t.Logf("res === %+v, args[3]==%d", res,args[3])
}



// 977 有序数组的平方
func sortedSquares(nums []int) []int {
	nl := len(nums)
	i, j := 0,nl-1
	res := make([]int, nl)
	for k := nl-1; k >= 0; k-- {
		tempI, tempJ := nums[i]*nums[i], nums[j]*nums[j]
		if tempI > tempJ {
			res[k] = tempI
			i++
		}else {
			res[k] = tempJ
			j--
		}
	}
	return res
}

func sortedSquares1(nums []int) []int {
	nl := len(nums)
	res := make([]int, 0, nl)
	lindex := 0
	for i := 0; i < nl && nums[i] < 0; i++ {
		lindex = i
	}
	for i, j := lindex, lindex+1; i >= 0 || j < nl; {
		if i < 0 {
			res = append(res, nums[j]*nums[j])
			j++
		}else if j == nl {
			res = append(res, nums[i]*nums[i])
			i--
		}else if nums[i]*nums[i] < nums[j]*nums[j] {
			res = append(res, nums[i]*nums[i])
			i--
		}else{
			res = append(res, nums[j]*nums[j])
			j++
		}
	}
	return res
}

// [-7,-3,2,3,11]
func Test977(t *testing.T) {
	args := []int{-7,-3,2,3,11}
	args1 := []int{-1,0,1,3,5,9,12}
	//t.Logf("search === %d", searchInsert1(args, 13))
	//t.Logf("res === %+v", sortedSquares(args))
	t.Logf("res === %+v", sortedSquares(args))
	t.Logf("res === %+v", sortedSquares(args1))
}

// 189 旋转数组
/*
	输入: nums = [1,2,3,4,5,6,7], k = 3
    输出: [5,6,7,1,2,3,4]
*/
func rotate0(nums []int, k int) []int { // leecode不给用？？？？？
	nl := len(nums)
	res := nums[nl-k:]
	res = append(res, nums[:nl-k]...)
	return res
}

func rotate(nums []int, k int) (res []int) {
	nl := len(nums)
	res = nums[nl-k:]
	res = append(res, nums[:nl-k]...)
	return
}

// [-7,-3,2,3,11][1,2,3,4,5,6,7]
//3
func Test189(t *testing.T) {
	args := []int{-7,-3,2,3,11}
	args1 := []int{1,2,3,4,5,6,7}
	t.Logf("res === %+v", rotate(args, 2))
	t.Logf("res === %+v", rotate(args1, 3))
}



// 283移动0
func moveZeroes(nums []int) []int {
	nl := len(nums)
	it := nl - 1
	nzero := nl - 1
	for ;it >= 0; it--{
		if nums[it] == 0 {
			nums = swapZero(nums, it, nzero)
			nzero--
		}
	}
	return nums
}

func swapZero(nums []int, i, j int) []int {
	pre := 0
	for p:=i; p < j; p++ {
		pre = nums[p]
		nums[p] = nums[p+1]
		nums[p+1] = pre
	}
	return nums
}

// 283移动0
func moveZeroes1(nums []int) []int {
	nl := len(nums)
	i, j := 0,0
	for ; i < nl; i++ {
		if nums[i] != 0 {
			nums[j] = nums[i]
			j++
		}
	}
	for j < nl {
		nums[j] = 0
		j++
	}
	return nums
}

// 283移动0
func moveZeroes2(nums []int) []int {
	nl := len(nums)
	i, j := 0,0
	for ; i < nl; i++ {
		if nums[i] != 0 {
			nums[j],nums[i] = nums[i],nums[j]
			j++
		}
	}
	return nums
}

func Test283(t *testing.T) {
	//args := []int{0,-7,-3,0,0,2,0,3,11}
	//t.Logf("res === %+v", moveZeroes2(args))
	t.Logf("res === %d", 4 >> 1)
}



// 167 两数之和II - 输入有序数组
func twoSum(numbers []int, target int) []int {
	nl := len(numbers)
	res := make([]int, 0, 2)
	i := 0
	for ; i < nl; i++ {
		search := halfSearch(numbers[i:], target - numbers[i])
		if search != -1 {
			res = append(res, i+1)
			res = append(res, search+i+1)
		}
	}
	return res
}

// 二分查找
func halfSearch(nums []int, target int) int {
	begin, end := 0, len(nums)
	for begin <= end {
		mid := (end + begin) >> 1
		if nums[mid] > target {
			end = mid-1
		}else if nums[mid] < target {
			begin = mid+1
		}else {
			return mid
		}
	}
	return -1
}

func Test167(t *testing.T) {
	args := []int{0,-7,-3,0,0,2,0,3,11}
	t.Logf("res === %+v", twoSum(args, 14))
}


// 344 反转字符串
func reverseString(s []byte) []byte {
	start, end := 0, len(s)-1
	for start < end {
		s[start], s[end] = s[end], s[start]
		start++
		end--
	}
	return s
}

func Test344(t *testing.T) {
	s := "asdf fd ss"
	args := []byte(s)
	t.Logf("res === %+v", args)
	t.Logf("res === %v", reverseString(args))
}



// 557 反转字符串中的单词
func reverseWords(s string) string {
	nums:=[]byte(s)
	nl := len(nums)
	low, high := 0, 0
	for ; high<nl;high++ {
		if nums[high] == ' ' {
			doReverse(nums, low, high-1)
			low = high+1
		}
		if high == nl-1 {
			doReverse(nums, low, high)
			low = high
		}
	}
	return string(nums)
}

func doReverse(s []byte, start, end int) []byte {
	for start < end {
		s[start], s[end] = s[end], s[start]
		start++
		end--
	}
	return s
}

func Test557(t *testing.T) {
	s := "asdf fd ss"
	args := []byte(s)
	t.Logf("res === %+v", args)
	t.Logf("res === %v", reverseWords(s))
}



// 删除元素
func removeElement(nums []int, val int) int {
	nl := len(nums)
	i, j := nl-1, nl-1
	for ;i >= 0; i-- {
		if nums[i] == val {
			nums[i], nums[j] = nums[j],nums[i]
			j--
		}
	}
	return j+1
}

func TestRemoveElement(t *testing.T) {
	args := []int{0,-7,-3,0,0,2,0,3,11}
	t.Logf("res === %+v", args)
	t.Logf("res === %d", removeElement(args, 15))
}


// 80
func removeDuplicatesII(nums []int) int {
	nl := len(nums)
	pre, curr := 0, 0
	cnt := 0
	res := 0
	for ;curr < nl; curr++ {
		if nums[curr] == nums[pre] {
			cnt++
		}
		if cnt <= 2 {
			res++
		}
		if nums[curr] != nums[pre] {
			pre = curr
			cnt = 0
		}
	}
	return res
}


func Test80(t *testing.T) {
	args := []int{1,1,1,2,2,3}
	t.Logf("res === %+v", args)
	t.Logf("res === %d", removeDuplicatesII(args))
}


type ListNode struct {
	Val int
	Next *ListNode
}

// 876. 链表的中间结点
func middleNode(head *ListNode) *ListNode {
	slow := head
	fast := head
	for fast != nil && fast.Next != nil {
		fast = fast.Next.Next
		slow = slow.Next
	}
	return slow
}


func middleNode1(head *ListNode) *ListNode {
	node := head
	res := head
	cnt := 0
	for node != nil {
		node = node.Next
		cnt++
	}
	for i := 0; i < cnt >> 1 + 1; i++ {
		res = res.Next
	}
	return res
}

func Test876(t *testing.T) {
	args := []int{1,1,1,2,2,3}
	t.Logf("res === %+v", args)
}


type pos struct {
	x int
	y int
}
func updateMatrix(matrix [][]int) [][]int {
	queue := make([]pos,0)
	nr,nc := len(matrix),len(matrix[0])
	for i:=0;i<nr;i++ {
		for j:=0;j<nc;j++ {
			if matrix[i][j] == 0 {
				queue = append(queue, pos{i,j})
			}else {
				matrix[i][j] = -1
			}
		}
	}
	directions := []pos{{-1,0},{0,1},{1,0},{0,-1}}
	for len(queue) >0 {
		cur := queue[0]
		queue = queue[1:]
		for _,dir := range directions {
			x := cur.x+dir.x
			y := cur.y+dir.y
			if x>=0 && x<nr && y>=0 && y<nc && matrix[x][y]==-1 {
				matrix[x][y] = matrix[cur.x][cur.y]+1
				queue = append(queue, pos{x,y})
			}
		}
	}
	return matrix
}


// 70 爬楼梯 （超时）
func climbStairs(n int) int {
	if n < 3 {
		return n
	}
	return climbStairs(n - 2) + climbStairs(n-1)
}
// 70 爬楼梯
func climbStairs1(n int) int {
	if n < 3 {
		return n
	}
	stp1 := 1
	stp2 := 2
	for i := 3; i<=n; i++ {
		temp := stp1 + stp2
		stp1 = stp2
		stp2 = temp
	}
	return stp2
}

// 198 打家劫舍
func rob(nums []int) int {
	nl := len(nums)
	if nl == 0 {
		return 0
	}
	if nl == 1 {
		return nums[0]
	}

	first := nums[0]
	second := max(nums[0], nums[1])


	for i := 2; i < nl; i++ {
		first, second = second, max(second, first+nums[i])
	}
	return second
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}


// 120 三角形最小路径和
func minimumTotal(triangle [][]int) int {
	n := len(triangle)
	f := [2][]int{}
	for i := 0; i < 2; i++ {
		f[i] = make([]int, n)
	}
	f[0][0] = triangle[0][0]
	for i := 1; i < n; i++ {
		curr := i % 2
		prev := 1 - curr
		f[curr][0] = f[prev][0] + triangle[i][0]
		for j := 1; j < i; j++ {
			f[curr][j] = min(f[prev][j - 1], f[prev][j]) + triangle[i][j]
		}
		f[curr][i] = f[prev][i - 1] + triangle[i][i]
	}
	ans := math.MaxInt32
	for i := 0; i < n; i++ {
		ans = min(ans, f[(n-1)%2][i])
	}
	return ans
}
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func Test231(t *testing.T) {
	//n := 1073741825
	t.Logf("res == %v", 1 << 30)
}
// 231
func isPowerOfTwo(n int) bool {
	if n == 0 {
		return false
	}
	if n == 1 {
		return true
	}
	if n%2 != 0 {
		return false
	}
	return isPowerOfTwo(n/2)
}
func isPowerOfTwo1(n int) bool {
	return n > 0 && n&(n-1) == 0
}
func isPowerOfTwo2(n int) bool {
	big := 1 << 30
	return n > 0 && big%n == 0
}


// 191 位1的个数
func hammingWeight(num uint32) (ones int) {
	for ; num > 0; num &= num - 1 {
		//fmt.Printf("num == %d \n", num)
		ones++
	}
	return
}
func hammingWeight1(num uint32) (ones int) {
	for i := 0; i < 32; i++ {
		if 1<<i&num > 0 {
			ones++
		}
	}
	return
}
func Test191(t *testing.T) {
	//n := 1073741825
	t.Logf("res == %v", hammingWeight(000000000000000000000010001011))
}

// 55 跳跃游戏
func canJump(nums []int) bool {
	nl := len(nums)
	if nl <= 1 {
		return true
	}
	dp := make([]bool, nl)
	dp[0] = true
	for i := 1; i < nl; i++ {
		for j := i - 1; j >= 0; j-- {
			if dp[j] && nums[j] + j >= i {
				dp[i] = true
				break
			}
		}
	}
	return dp[nl-1]
}

func canJump1(nums []int) bool {
	nl := len(nums)
	if nl <= 2 {
		return true
	}
	maxJump := 0
	for i := 0; i < nl; i++ {
		if i <= maxJump {
			maxJump = max55(maxJump, i + nums[i])
			if maxJump >= nl-1 {
				return true
			}
		}
	}
	return false
}

func max55(a,b int) int {
	if a >= b {
		return a
	}
	return b
}
func Test55(t *testing.T) {
	nums := []int{2,3,1,1,4}
	//nums := []int{3,2,1,0,4}
	//t.Logf("test55 res === %v \n", canJump(nums))
	t.Logf("test55 res === %v \n", canJump1(nums))
}

// 62 不同路径
func uniquePaths(m int, n int) int {
	dp := make([][]int, m)
	for i := range dp {
		dp[i] = make([]int, n)
		dp[i][0] = 1
	}
	for i := 0; i < n; i++ {
		dp[0][i] = 1
	}
	for i := 1; i < m; i++ {
		for j := 1; j < n; j++ {
			dp[i][j] = dp[i-1][j] + dp[i][j-1]
		}
	}
	return dp[m-1][n-1]
}
// 62 滚动数组优化
func uniquePaths1(m int, n int) int {
	if n == 1 || m == 1 {
		return 1
	}
	if m < n {
		m,n = n,m
	}
	dp := make([]int, n)
	for i := 0; i < n; i++ {
		dp[i] = 1
	}
	for i := 1; i < m; i++ {
		for j := 1; j < n; j++ {
			dp[j] += dp[j-1]
		}
	}
	return dp[n-1]
}
// 62 组合数学
func uniquePaths2(m, n int) int {
	return int(new(big.Int).Binomial(int64(m+n-2), int64(n-1)).Int64())
}
func Test62(t *testing.T) {

}

// 64 z最小路径和
func minPathSum(grid [][]int) int {
	l := len(grid)
	if l == 0 {
		return 0
	}
	for i := 0; i < l; i++ {
		for j := 0; j < len(grid[i]); j++ {
			if i == 0 && j != 0 {
				grid[i][j] = grid[i][j-1] + grid[i][j]
			} else if j == 0 && i != 0 {
				grid[i][j] = grid[i-1][j] + grid[i][j]
			} else if i !=0 && j != 0 {
				grid[i][j] = min64(grid[i-1][j], grid[i][j-1]) + grid[i][j]
			}
		}
	}
	return grid[l-1][len(grid[l-1])-1]
}
func min64(x, y int) int {
	if x < y {
		return x
	}
	return y
}
func Test64(t *testing.T) {

}

// 63 不同路径2
func uniquePathsWithObstacles(obstacleGrid [][]int) int {
	if len(obstacleGrid) == 0 || len(obstacleGrid[0]) == 0 {
		return 0
	}

	n, m := len(obstacleGrid), len(obstacleGrid[0])
	f := make([]int, m)
	if obstacleGrid[0][0] == 0 {
		f[0] = 1
	}
	for i := 0; i < n; i++ {
		for j := 0; j < m; j++ {
			if obstacleGrid[i][j] == 1 {
				f[j] = 0
				continue
			}
			if j - 1 >= 0 && obstacleGrid[i][j-1] == 0 {
				f[j] += f[j-1]
			}
		}
	}
	return f[len(f)-1]
}
func Test63(t *testing.T) {

}

// 45 跳跃游戏2
func jump(nums []int) int {
	nl := len(nums)
	if nl < 2  {
		return 0
	}
	res := 1
	for i := 0; i < nl; {
		if i+nums[i] > nl-1 {
			break
		}
		res += 1
		maxChoose := nums[i] + i
		for j := i+1; j <= i+nums[i]; j++ {
			if maxChoose < nums[j] + j {
				maxChoose = nums[j] + j
			}
		}
		if maxChoose >= nl-1 {
			break
		}
		i = maxChoose
	}
	return res
}
func jump1(nums []int) int {
	nl := len(nums)
	end := 0
	maxPosition := 0
	steps := 0
	for i := 0; i < nl - 1; i++ {
		if maxPosition < i+nums[i] {
			maxPosition = i+nums[i]
		}
		if maxPosition >= nl-1 {
			steps++
			break
		}
		if i == end {
			end = maxPosition
			steps++
		}
	}
	return steps
}
func Test45(t *testing.T) {
	nums := []int{1,2,3}
	t.Logf("res == %d \n", jump(nums))
	// f(0) = 0
	// f(i) = f(i-1) + f(i)
}

// 34. 在排序数组中查找元素的第一个和最后一个位置
func searchRange(nums []int, target int) []int {
	nl := len(nums)
	start, end := 0, nl-1
	resStart, resEnd, index := 0,0,-1
	for start <= end {
		mid := (start+end) >> 1
		if nums[mid] > target {
			end = mid-1
		} else if nums[mid] < target{
			start = mid + 1
		}else {
			index = mid
			break
		}
	}
	if index == -1 {
		return []int{-1,-1}
	}
	resStart = index
	for i:=index-1;i>=0;i-- {
		if nums[i] != target {
			break
		}
		resStart--
	}
	resEnd = index
	for i:=index+1;i<nl;i++ {
		if nums[i] != target {
			break
		}
		resEnd++
	}
	return []int{resStart, resEnd}
}
func searchRange1(nums []int, target int) []int {
	leftmost := sort.SearchInts(nums, target)
	if leftmost == len(nums) || nums[leftmost] != target {
		return []int{-1, -1}
	}
	rightmost := sort.SearchInts(nums, target + 1) - 1
	return []int{leftmost, rightmost}
}

func Test34(t *testing.T) {
	nums := []int{1}
	t.Logf("res == %v \n", searchRange(nums, 1))
	t.Logf("res == %v \n", (0+0) >> 1)
}

// 33. 搜索旋转排序数组
func search33(nums []int, target int) int {
	nl := len(nums)
	start, end := 0, nl-1
	index := -1
	for start <= end {
		mid := (start + end) >> 1
		if nums[mid] == target {
			index = mid
			break
		}else if nums[start] < nums[mid] {
			if nums[mid] < target {
				end = mid - 1
			}else {
				start = mid + 1
			}
		}else {
			index = mid
			break
		}
	}
	return index
}
func Test33(t *testing.T) {
	nums := []int{4,5,6,7,0,1,2}
	t.Logf("res == %v \n", search33(nums, 0))
}


// 153. 寻找旋转排序数组中的最小值
func findMin(nums []int) int {
	nl := len(nums)
	start, end := 0, nl-1
	for start <= end {
		if end - start == 1 && nums[end] < nums[start] {
			return nums[end]
		}
		mid := (start + end) >> 1
		if nums[mid] > nums[start] && nums[mid] > nums[end] {
			start = mid
		}else if nums[mid] < nums[end]{
			end = mid
		}else {
			return nums[start]
		}
	}
	return nums[0]
}

// 153. 寻找旋转排序数组中的最小值 优化
func findMin1(nums []int) int {
	nl := len(nums)
	start, end := 0, nl-1
	for start < end {
		mid := start + (end - start) >> 1
		if nums[mid] < nums[end]{
			end = mid
		}else {
			start = mid+1
		}
	}
	return nums[start]
}
func Test153(t *testing.T) {
	//nums := []int{4,5,6,0,1,2}
	nums := []int{11,13,15,17}
	t.Logf("res == %v \n", findMin1(nums))
}


// 162. 寻找峰值
func findPeakElement(nums []int) int {
	nl := len(nums)
	start, end := 0, nl-1
	for start < end {
		mid := start + (end - start) >> 1
		if nums[mid] < nums[end]{
			end = mid
		}else {
			start = mid+1
		}
	}
	return nums[start]
}
func Test162(t *testing.T) {
	//nums := []int{4,5,6,0,1,2}
	nums := []int{11,13,15,17}
	t.Logf("res == %v \n", findPeakElement(nums))
}


// 438. 找到字符串中所有字母异位词
func findAnagrams(s string, p string) []int {
	sl, pl := len(s), len(p)
	if sl < pl {
		 return nil
	}
	cnts, cntp := [26]int{}, [26]int{}
	for i,_:=range p{
		cntp[p[i]-'a']++
	}
	res := make([]int,0)
	l,r := 0,0
	for ; r < sl; r++ {
		cnts[s[r]-'a']++
		if r-l+1 >= pl {
			if cnts == cntp {
				res = append(res, l)
			}
			cnts[s[l]-'a']--
			l++
		}
	}
	return res
}
func Test438(t *testing.T) {
	s := "gqmmqg"
	p := "gqm"
	t.Logf("sl === %v \n", findAnagrams(s,p))
}


// 200 岛屿数量 BFS
func numIslands(grid [][]byte) int {
	r := len(grid)
	if r == 0 {
		return 0
	}
	c := len(grid[0])
	count := 0
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			if grid[i][j] == '1' {
				BFS(grid,i,j)
				count++
			}
		}
	}
	return count
}
func BFS(grid [][]byte, i, j int) {
	dx:=[]int{-1,1,0,0}
	dy:=[]int{0,0,1,-1}
	queue := make([]int, 0)
	queue = append(queue, i, j)
	grid[i][j] = '0'
	for len(queue) != 0 {
		i,j := queue[0], queue[1]
		queue = queue[2:]
		for p := 0; p < 4; p++ {
			tempI := i+dx[p]
			tempJ := j+dy[p]
			if tempI >= 0 && tempI < len(grid) && tempJ >= 0 && tempJ < len(grid[0]) && grid[tempI][tempJ] == '1' {
				grid[tempI][tempJ] = '0'
				queue = append(queue, tempI, tempJ)
			}
		}
	}
}
// 200 DFS
var dx = [4]int{-1, 1, 0, 0}
var dy = [4]int{0, 0, 1, -1}
func numIslands1(grid [][]byte) int {
	m := len(grid)
	if m == 0{
		return 0
	}

	n := len(grid[0])
	count := 0
	for i:=0; i<m; i++{
		for j:=0; j<n; j++{
			if grid[i][j] == '1'{
				DFS(grid, i, j)
				count++
			}
		}
	}
	return count
}
func DFS(grid [][]byte, i, j int){
	if i<0 || j<0 || i >= len(grid) || j >=len(grid[0]) || grid[i][j] != '1'{
		return
	}

	grid[i][j] = '0'
	for k:=0; k < 4; k++{
		DFS(grid, i+dx[k], j+dy[k])
	}
}


// 200 并查集   使用路径压缩和rank排名
var row, col int
var parent []int
var count int
var rank []int
func numIslands2(grid [][]byte) int {
	row = len(grid)
	if row == 0{
		return 0
	}
	col = len(grid[0])
	//path compress
	count = 0
	parent = make([]int, row*col)
	rank = make([]int, row*col)
	for i:=0; i<row; i++{
		for j:=0; j<col; j++{
			if grid[i][j] == '1'{
				parent[i*col+j] = i*col + j
				count++
			}
		}
	}
	for i:=0; i<row; i++{
		for j:=0; j<col; j++{
			if grid[i][j] == '1'{
				for k:=0; k<4; k++{
					nr, nc := i+dx[k], j+dy[k]
					if isValid(grid, nr, nc) && grid[nr][nc] == '1'{
						union(i*col + j, nr*col + nc)
					}
				}
			}
		}
	}
	return count
}
func isValid(grid [][]byte, i, j int)bool{
	return !(i < 0 || i >= row || j <0 || j>=col)
}
func find(i int)int{
	if parent[i] != i{
		parent[i] = find(parent[i])
	}
	return parent[i]
}
func union(x, y int){
	rootx := find(x)
	rooty := find(y)

	//rank
	if rootx != rooty{
		if rank[rootx] > rank[rooty]{
			parent[rooty] = rootx
		}else if rank[rootx] < rank[rooty]{
			parent[rootx] = rooty
		}else{
			parent[rooty] = rootx
			rank[rootx] += 1
		}
		count -= 1
	}

}

//  Definition for a Node.
type Node struct {
	Val int
	Left *Node
	Right *Node
	Next *Node
}
// 117. 填充每个节点的下一个右侧节点指针 II
func connect(root *Node) *Node {
	if root == nil {
		return root
	}
	que := []*Node{root}
	for len(que) > 0 {
		temp := que
		que = nil
		for i, node := range temp {
			if i+1 < len(temp) {
				node.Next = temp[i+1]
			}
			if node.Left != nil {
				que = append(que, node.Left)
			}
			if node.Right != nil {
				que = append(que, node.Right)
			}
		}
	}
	return root
}
func connect1(root *Node) *Node {
	start := root
	for start != nil {
		var nextStart, last *Node
		handle := func(cur *Node) {
			if cur == nil {
				return
			}
			if nextStart == nil {
				nextStart = cur
			}
			if last != nil {
				last.Next = cur
			}
			last = cur
		}
		for p := start; p != nil; p = p.Next {
			handle(p.Left)
			handle(p.Right)
		}
		start = nextStart
	}
	return root
}


// 1091. 二进制矩阵中的最短路径
type table struct {
	r,c,d int
}
func shortestPathBinaryMatrix(grid [][]int) int {
	if len(grid) == 0 || len(grid[0]) == 0 || grid[0][0] == 1 {
		return -1
	}
	dict := [][]int{{1,-1}, {1,0}, {1,1}, {0,-1}, {0,1}, {-1, -1}, {-1,0}, {-1, 1}}
	r,c := len(grid), len(grid[0])
	var que []table
	que = append(que, table{0, 0, 1})
	for len(que) > 0 {
		cur := que[0]
		que = que[1:]
		if cur.r == r-1 && cur.c == c-1 { // 最早出现，则说明是最短路径
			return cur.d
		}
		for _, d := range dict {
			nr := cur.r+d[0]
			nc := cur.c+d[1]
			if nr >= 0 && nr <= r-1 && nc >= 0 && nc <= c-1 && grid[nr][nc] == 0 {
				que = append(que, table{nr, nc, cur.d+1})
				grid[nr][nc] = 1
			}
		}
	}
	return -1
}



func Test96(t *testing.T) {
	t.Logf("numTrees : %d \n", numTrees(3))
}
// 96 不同的二叉搜索树
func numTrees(n int) int {
	C := 1
	for i := 0; i < n; i++ {
		C = C * 2 * (2 * i + 1) / (i + 2)
	}
	return C
}

func Test2222(t *testing.T) {
	a := 2
	b := 3
	t.Logf("res = %d", int32(math.Abs(float64(a-b))))
}



// 剑指 Offer 10- I. 斐波那契数列
func fib(n int) int {
	//const mod int = 1e9 + 7
	if n < 2 {
		return n
	}

	p, q, r := 0, 0, 1
	for i := 2; i <= n; i++ {
		p = q
		q = r
		//r = (q+p) % mod
		r = q+p
	}
	return r
}

func TestOffer101(t *testing.T) {
	t.Logf("res = %d", waysToStep(45))
	//const mod int = 1e9 + 7
	//t.Logf("res222 = %v", 101 % mod)
}
// 面试题 08.01. 三步问题
func waysToStep(n int) int {
	if n == 1 || n == 2 {
		return n
	}
	if n == 3 {
		return 4
	}
	dp1, dp2, dp3 := 1, 2, 4
	for i := 4; i <= n; i++ {
		dp1, dp2, dp3 = dp2, dp3, (dp1+dp2+dp3) % 1000000007
	}
	return dp3
}


// 46. 使用最小花费爬楼梯
func minCostClimbingStairs(cost []int) int {
	cl := len(cost)
	if cl == 0 {
		return 0
	}
	if cl < 2 {
		return cost[0]
	}
	pre, curr := 0, 0
	for i := 2; i <= cl; i++ {
		pre , curr = curr, min46(curr + cost[i-1], pre + cost[i-2])
	}
	return curr
}
func min46(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func Test46(t *testing.T) {
	num := []int{100}
	res := minCostClimbingStairs(num)
	t.Logf("res = %d, time = %d \n", res, int32(time.Now().Unix()))
}



// 122. 买卖股票的最佳时机 II

//func maxProfit(prices []int) int {
//	if len(prices) < 2 {
//		return 0
//	}
//	pl := len(prices)
//	pre, curr, sum := 1,0, 0
//	flag := false
//	for ; pre < pl; pre++ {
//		if flag == false && curr < pre {
//			flag = true
//		}
//		if flag == true && curr < pre {
//
//		}
//	}
//}
//
//func Test122(t *testing.T) {
//
//}


// 918. 环形子数组的最大和
// 首先考虑这不是环形数组，那么可以用动态规划很容易的计算出子数组的最大和
//maxNow = Max(A[i], maxNow + A[i]) // 计算直到A[i]的最大值
//max = Max(max, maxNow)
////  eg: 5 -4 3, i=0, maxNow = 5, max = 5
//// 		i=1, maxNow = 1, max = 5
//// 		i=2, maxNow = 4, max = 5
//
//再考虑环形数组 最大值有可能是去掉中间一段负数，那么就可以转换为计算非环形数组的最小值。
//minNow = Min(A[i], minNow + A[i])
//min = Min(min, minNow)
//// 那么可能的最大值就是 total - min
//max1 = total - min
//特殊考虑数组最大值 < 0的情况
//
//完整代码
func maxSubarraySumCircular(A []int) int {
	max, maxNow := -2 << 31, -2 << 31
	min, minNow := 2 << 31, 2 << 31
	maxVal := -2 << 31
	total := 0
	for i := 0; i < len(A); i++ {
		total += A[i]
		maxVal = max918(maxVal, A[i])

		maxNow = max918(A[i], maxNow + A[i])
		max = max918(max, maxNow)

		minNow = min918(A[i], minNow + A[i])
		min = min918(min, minNow)
	}
	if maxVal <= 0 {
		return maxVal
	}

	return max918(max, total - min)
}
func max918(x, y int) int {
	if x > y {
		return x
	}
	return y
}
func min918(x, y int) int {
	if x < y {
		return x
	}
	return y
}
func Test918(t *testing.T) {
	//num := []int{1,-2,3,-2}
	//num := []int{5,-3,5}
	//num := []int{3,-1,2,-1}
	//num := []int{3,-2,2,-3}
	num := []int{-2,-3,-1}
	t.Logf("res = %v ", maxSubarraySumCircular(num))
}


// 152. 乘积最大子数组
func maxProduct(nums []int) int {
	fmax, fmin, res, nl := nums[0], nums[0], nums[0], len(nums)
	for i := 1; i < nl; i++ {
		tmax, tmin := fmax, fmin
		fmax = max152(tmax*nums[i], max152(tmin*nums[i],nums[i]))
		fmin = min152(tmin*nums[i], min152(tmax*nums[i], nums[i]))
		res = max(fmax, res)
	}
	return res
}
func max152(x, y int) int {
	if x > y {
		return x
	}
	return y
}
func min152(x, y int) int {
	if x < y {
		return x
	}
	return y
}


func Test1567(t *testing.T) {
	//num := []int{1,-2,3,-2}
	//num := []int{5,-3,5}
	//num := []int{3,-1,2,-1}
	num := []int{3,-2,2,-3}
	t.Logf("res = %v ", getMaxLen(num))
}
// 1567. 乘积为正数的最长子数组长度 动态规划
func getMaxLen(nums []int) int {
	//包含nums[i]的最大值
	maxdp:=make([]int,len(nums)+1)
	//包含nums[i]的最小值
	mindp:=make([]int,len(nums)+1)
	max:=0
	// 遍历
	for i:=1;i<=len(nums);i++{
		//如果当前值是0
		//那么包含当前值的最大最小都是0
		if nums[i-1]==0{
			maxdp[i],mindp[i]=0,0
			// 如果当前值大于0
		}else if nums[i-1]>0{
			//最大值等于前一个最大值+1
			maxdp[i]=maxdp[i-1] + 1
			// 如果前一个最小值不等于0 最小值等于前一个加一
			if mindp[i-1]!=0{
				mindp[i]=mindp[i-1]+1
			}
		}else{
			//如果当前值小于0
			// 如果前一个最小值大于0，那么最大值就是前一个最小值+1
			if mindp[i-1]>0{
				maxdp[i]=mindp[i-1]+1
			}
			// 最小值等于 最大值+1
			mindp[i]=maxdp[i-1]+1
		}
		// 记录最大的
		if maxdp[i]>max{
			max=maxdp[i]
		}
	}
	return max
}

func Test102(t *testing.T) {
	var root = &TreeNode{
		Val:1,
		Left: &TreeNode{2, &TreeNode{Val:3}, &TreeNode{Val:4, Right: &TreeNode{Val:5}}},
		Right: &TreeNode{6, &TreeNode{Val:7, Right: &TreeNode{Val:8}}, &TreeNode{Val:9}},
	}
	levelOrder(root)
	t.Logf("levelOrder ret == %+v", levelOrder(root))
}
// 102 层序遍历
func levelOrder(root *TreeNode) (ret [][]int) {
	ret = [][]int{}
	if root == nil {
		return
	}
	var q = []*TreeNode{root}
	var i int32
	for i = 0; len(q) > 0; i ++ {
		ret = append(ret, []int{})
		var nextLevel []*TreeNode
		var j int
		for j = 0; j < len(q); j++ {
			var node = q[j]
			ret[i] = append(ret[i], node.Val)

			if node.Left != nil {
				nextLevel = append(nextLevel, node.Left)
			}
			if node.Right != nil {
				nextLevel = append(nextLevel, node.Right)
			}
		}
		q = nextLevel
	}
	return ret
}

// 107. 二叉树的层序遍历 II
//func levelOrderBottom(root *TreeNode) [][]int {
//
//}


// 95 不同的二叉搜索树II
func generateTrees(n int) []*TreeNode {
	if n == 0 {
		return nil
	}
	return searchTree(1, n)
}
func Test95(t *testing.T) {
	for _, tree := range generateTrees(3) {
		t.Logf("generateTrees ret == %+v \n", tree)
	}
}
func searchTree(start, end int) (st []*TreeNode) {
	st = []*TreeNode{}
	if start > end {
		st = append(st, nil)
		return
	}
	for i := start; i <= end; i++ {
		var leftTrees = searchTree(start, i-1)
		var rightTree = searchTree(i+1, end)

		for _, left := range leftTrees {
			for _, right := range rightTree {
				st = append(st, &TreeNode{i, left, right})
			}
		}
	}
	return
}
