-module(test).
-author({ipso, backend}).
-define('PI', 3.1415).
-purpose("test erlang").
-export([
    start/0, 
    get_cats/1, 
    perms/1,
    get_circle_area/1,
    my_file_rename/1,
    open_and_write/1,
    println_map/1,
    to_term_test/1,
    two_list/1,
    test_line/0,
    find_value/2,
    yaojiu_and_zipai_md1/1,
    yaojiu_and_zipai_md2/1,
    test/4,
    is_single/1,
    print_single/1,
    print_common/2,
    print_special/2
]).


%% =======================================
%% 我的erlang入门小练习，刚开始受其他语言的影响
%% 会特别不适应erlang的基础语法，大量的练习是非
%% 常有必要的
%% =======================================


% ====================================== practice 1 ======================================== %
%% 根据sdk文档自由练习使用module


%% =======================================
%% 使用?MODULE来获取module名称
%% @Return echo
%% @end
start() -> io:format("~p~n", [?MODULE]).

%% =======================================
%% 获取值为0到N的列表
%% @Return list
get_cats(0) -> [0];
get_cats(N) ->     
    [N | get_cats(N-1)].

%% =======================================
%% perms函数，获取列表的所有元素的自由组合列表
%% Return list
perms([]) -> [[]];
perms(L) -> [[H|T] || H <- L, T <- perms(L -- [H])].

%% =======================================
%% 通过半径求得圆得面积
%% @Return number() 圆面积
get_circle_area(Radius) ->
    2 * Radius * Radius * ?PI.


%% =======================================
%% 修改文件名
%% Return  ok | Error
my_file_rename(NewName) when is_list(NewName) -> 
    file:rename("erl.hrl", NewName);
my_file_rename(NotFile) ->
    {error, "请输入正确得文件名！~p~n", [NotFile]}.


%% =======================================
%% 打开文件并向文件写入数据
%% Return  ok | Error
open_and_write(FileName) when is_list(FileName) ->
    Data = "{ipso, 'backend', 22}.",
    {ok, Fd} = file:open(FileName, [write, raw]),
    file:write(Fd, list_to_binary(Data)),
    file:close(Fd);
open_and_write(ErrorName) -> 
    {error, "Format Error: ", ErrorName}.

%% =======================================
%% 打印map
%% return   ok | error
println_map(Name) -> 
    MyMaps = #{name => "ipso", age  => 22},
    Maps1 = MyMaps#{name := Name},
    io:format("MyMaps:~p~nMaps1:~p~n", [MyMaps, Maps1]).



%% TestParam <<131,104,3,100,0,1,97,100,0,1,98,100,0,1,99>>, term_to_binary([a, b, c]
to_term_test(Param) -> 
    io:format("~p~n", [erlang:binary_to_term(Param)]).

%% TestParam  [11, 2, 2.3, 4]
two_list([]) -> [];
two_list([H | T]) -> 
    [ X || X <- [H], erlang:is_integer(X) == true] ++ two_list(T).


%% 打印?LINE(语句所在行数)
test_line() -> 
    io:format("NYI is: ~p ~p ~n", [?MODULE, ?LINE]).



% ====================================== practice 2 ======================================== %
%% 判断list有无幺九和2个以上的字牌
%% @Param L  卡牌list    TestParam [22, 23, 33, 41, 42, 43, 11]

%% 判断某个卡牌是否是幺九
is_yaojiu(Card) -> 
    YaoJiu = [11, 21, 31, 19, 29, 39],
    lists:member(Card, YaoJiu).

%% 判断某个卡牌是否是字牌
is_zipai(Card) -> 
    ZiPai = [41, 42, 43, 51, 52, 53, 54],
    lists:member(Card, ZiPai).

%% ===================================
%% 判断list有无幺九和2个以上的字牌
%% @Param L  卡牌list    TestParam [22, 23, 33, 41, 42, 43, 11]
yaojiu_and_zipai_md1(L) ->
    Len = erlang:length(L),
    case Len > 2 of
        false ->
            false;
        true -> 
            GetYaoJiu = [YJ || YJ <- L, is_yaojiu(YJ) =:= true],
            GetZiPai = [ZP || ZP <- L, is_zipai(ZP) =:= true],
            (erlang:length(GetYaoJiu) >= 1) and (erlang:length(GetZiPai) > 2)
    end.
    


%% ========================================
%% 从列表中查找某个值
%% @Param  V  term()   TestParam: integer()
%% @Param  [H|T]  H:列表头，T:列表尾部   TestParam: [11, 2, 23, 4, 33, 44, 11, 12]
%% @Return true | false 
find_value(_, []) -> false;
find_value(V, [H | T]) -> 
    if
        V == H ->
            true;
        V /= H ->
            find_value(V, T)
    end.

%% ========================================
%% 自定义foreach
-spec my_foreach(Fun, L1, L2, L3) -> [T] when
      Fun :: fun((Elem :: T) -> term()),
      L1 :: [T],
      L2 :: [T],
      L3 :: [T],
      T :: term().

my_foreach(_, [], _, Res) -> Res;
my_foreach(Fun, [H | T], L, Res) -> 
    my_foreach(Fun, T, L, Res ++ [X || X <- [H], Fun({X, L}) == true]).


%% =======================================
%% 判断list有无幺九和2个以上的字牌 方法2
%% @Param L  卡牌list   TestParam [22, 23, 33, 41, 42, 43, 11]
yaojiu_and_zipai_md2(L) -> 
    Len = erlang:length(L),
    case Len > 2 of
        false ->
            false;
        true -> 
            YaoJiu = [11, 21, 31, 19, 29, 39],
            ZiPai  = [41, 42, 43, 51, 52, 53, 54],
            F = (fun({Card, Cards}) -> find_value(Card, Cards) end),
            GetYaoJiu = my_foreach(F, L, YaoJiu),
            GetZiPai  = my_foreach(F, L, ZiPai),
            (erlang:length(GetYaoJiu) >= 1) and (erlang:length(GetZiPai) > 2)
    end.



% ====================================== practice 2 ======================================== %
%% 输入一个列表，输出3个以上X、Y、Z的数量，例如输入（[41,42,41,41,43,42,12,19,39]）
%% 那么输出为1，（[41,41,41,43,43,43]）输出为2

%% =========================================
%% 统计某个值在列表中的数量
-spec value_num(V, L1, Res) -> T when
      V :: T,
      L1 :: [T],
      Res :: T,
      T :: term().

value_num(_, [], Res) -> Res;
value_num(V, [H | T], Res) ->
    if
        V =:= H ->
            value_num(V, T, Res + 1);
        V =/= H -> 
            value_num(V, T, Res)
    end.

test(X, Y, Z, L) when erlang:length(L) > 2 -> 
    XN = case value_num(X, L, 0) > 2 of
        true ->
            1;
        false -> 
            0
    end,
    YN = case value_num(Y, L, 0) > 2 of
        true ->
            1;
        false -> 
            0
    end,
    ZN = case value_num(Z, L, 0) > 2 of
        true ->
            1;
        false -> 
            0
    end,
    XN + YN + ZN;

test(_X, _Y, _Z, _L) -> 
    0.



% ====================================== practice 3 ======================================== %
%% lists模块练习，lists模块在业务开发中属于常用模块，需要重点关注和练习该模块



% =================================  practice 3.1 ================================= %
%% 3、检测一个列表中是否存在单个元素
%% Sample Input  [1, 2]  [1, 1]  []  [1, 2, 1, 2, 3, 4]

is_single(L) -> 
    is_single(L, L).

is_single([], _L) -> false;
is_single([H|T], L) ->
    case value_num(H, L, 0) of
        Num when Num > 1 ->
            is_single(T, L);
        _ -> 
            true
    end.


% =================================  practice 3.2 ================================= %
%% 4、输入一个列表，输出所有的单个元素以及单个元素的个数
%% Sample Input  [1, 2]  [1, 1]  []  [1, 2, 1, 2, 3, 4]


print_single(L) -> 
    Res = print_single(L, L, []),
    io:format("列表中所有的单个元素: ~p~n", [Res]),
    io:format("列表中单个元素的个数: ~p~n", [erlang:length(Res)]).

print_single([], _L, Res) -> Res;
print_single([H|T], L, Res) ->
    NRes = case value_num(H, L, 0) of
        Num when Num > 1 ->
            [];
        _ -> 
            [H]
    end,
    print_single(T, L, Res ++ NRes).



% =================================  practice 3.3 ================================= %
%% 5、现有2个列表，输出二者共有的元素和独有的元素

print_common(L1, L2) -> 
    Res = lists:merge(L1, L2) -- lists:umerge(L1, L2),
    io:format("二者共有的元素: ~p~n", [Res]).


print_special(L1, L2) -> 
    Res = lists:umerge(L1, L2) -- (lists:merge(L1, L2) -- lists:umerge(L1, L2)),
    io:format("二者独有的元素: ~p~n", [Res]).




% ====================================== practice 4 ======================================== %
%% erlang调试程序，根据erlang的特性，在erlang中调试程序并不需要像Java或者C那样设置断点追踪栈中变量的变化情况，
%% java或者C他们的变量存储是可以绑定多个引用的，每个引用都可以对变量进行修改，所以需要追踪栈信息监视变量变化情况，
%% 但是在erlang中变量值所占内存是不允许多绑定的，所以对于变量值绑定后不再修改，无需再跟踪，因此erlang中的调试方法
%% 就是打印。Erlang程序员使用各种各样的方法来调试他们的程序。到目前为止，最常用的方法就是给有 问题的程序添加打印语
%% 句。但如果想查看的数据结构变得非常大，这种方法就无效了，在这种情 况下可以把它们转储到一个文件，留待将来检查。一些人
%% 使用错误记录器来保存错误消息，另一些人则把它们写入文件。如果都实现不了，还可以使用Erlang调试器或者跟踪程序的执行过程。


%% 如果感兴趣的数据结构很大，就可以把它写入一个文件, 可以使用以下做通用函数
dump(File, Term) ->
    Out = File ++ ".tmp",
    io:format("** dumping to ~s~n", [Out]),
    {ok, S} = file:open(Out, [write]),
    io:format(S, "~p.~n", [Term]),
    file:close(S).