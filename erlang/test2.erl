-module(test2).
-author({gqm, erlang_backend}).

-export([
    test6/0,
    test7/0,
    get_max/1,
    get_max/2,
    print_common/2,
    print_special/2,
    get_special/3,
    get_commmon/3,
    common_and_Sp/2,
    get_comm_and_sp/4
]).


% =================================  test  ================================= %

%% practice 6 参数测试
test6() -> 
    {error, "bad param"} = get_max([]),
    {error, "bad param"} = get_max(error_param),
    22 = get_max([1,22,4,3]),
    test6_pass.

%% practice 7 参数测试
test7() -> 
    {error,"bad param"} = print_common(error, ss),
    {error,"bad param"} = print_common([22,33], ss),
    {error,"bad param"} = print_common(error, [22,33]),
    test7_pass.

% =================================  practice 6 ================================= %
%% 6、设计一个自己的函数，求一个列表中的最大值，要求不用lists模块的函数

%% 获取最大值，接收参数，并进行参数验证函数
get_max(L) when erlang:is_list(L), erlang:length(L) > 0 ->
    get_max(L, 0);
get_max(_L) -> 
    {error, "bad param"}.

%% 获取最大值逻辑处理函数
get_max([], Res) -> Res;
get_max([H|T], Res) ->
    case H > Res of
        true ->
            get_max(T, H);
        false ->
            get_max(T, Res)
    end .


% =================================  practice 7-m1  ================================= %
%% 7、现有2个列表，输出二者共有的元素和独有的元素，并且去重，重新设计一下，要求只能用到lists:member

%% 共有元素函数
print_common(L1, L2) when erlang:is_list(L1), erlang:is_list(L2) -> 
    Res = get_commmon(L1, L2, []),
    NRes = change(Res, []),
    io:format("二者共有的元素: ~p~n", [NRes]);

print_common(_L1, _L2) -> 
    {error, "bad param"}.



%% 独有元素函数
print_special(L1, L2) when erlang:is_list(L1), erlang:is_list(L2) -> 
    L1Sp = get_special(L1, L2, []),
    L2Sp = get_special(L2, L1, []),
    Res = L1Sp ++ L2Sp,
    NRes = change(Res, []),
    io:format("二者独有的元素: ~p~n", [NRes]);

print_special(_L1, _L2) -> 
    {error, "bad param"}.



% =================================  practice 7-m2  ================================= %
%% 7、现有2个列表，输出二者共有的元素和独有的元素，并且去重，重新设计一下，要求只能用到lists:member

%% 共有元素和独有元素函数
common_and_Sp(L1, L2) when erlang:is_list(L1), erlang:is_list(L2) -> 
    {Comm, SpL1} = get_comm_and_sp(L1, L2, [],[]),
    {_, SpL2} = get_comm_and_sp(L2, L1, [],[]),
    Sp = SpL1 ++ SpL2,
    NComm = change(Comm, []),
    NSp = change(Sp, []),
    {{common, NComm}, {special, NSp}};

common_and_Sp(_L1, _L2) -> 
    {error, "bad param"}.



% =================================  tool fun  ================================= %

%% 返回在第一个list不在第二个list中的元素的列表
%% 辅助函数
get_special([], _L, Res) -> Res;
get_special([H|T], L, Res) -> 
    case lists:member(H, L) of
        false ->
            get_special(T, L, [H|Res]);
        true -> 
            get_special(T, L, Res)
    end.


%% 返回两个列表的共有元素
%% 辅助函数
get_commmon([], _L, Res) -> Res;
get_commmon([H|T], L, Res) -> 
    case lists:member(H, L) of
        true ->
            get_commmon(T, L, [H|Res]);
        false -> 
            get_commmon(T, L, Res)
    end.
    

%% 去重函数
%% 辅助函数
change([], Res) -> Res;
change([H|T], Res) -> 
    case lists:member(H, Res) of
        true -> 
            change(T, Res);
        false ->
            change(T, [H|Res])
    end. 


%% 返回两个列表的共有元素
%% 辅助函数
get_comm_and_sp([], L, Comm, Sp) -> {Comm, Sp};
get_comm_and_sp([H|T], L, Comm, Sp) -> 
    case lists:member(H, L) of
        true ->
            get_comm_and_sp(T, L, [H|Comm], Sp);
        false -> 
            get_comm_and_sp(T, L, Comm, [H|Sp])
    end.