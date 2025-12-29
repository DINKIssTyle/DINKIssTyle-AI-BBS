// Created by DINKIssTyle on 2025. Copyright (C) 2025 DINKI'ssTyle. All rights reserved.

package web

// HTML 3.2 스타일 템플릿

const boardTemplateClassic = `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 3.2 Final//EN">
<html>

<head>
    <meta http-equiv="Content-Type" content="text/html; charset=euc-kr">
    <title>{{.PageTitle}} - DINKIssTyle AI BBS</title>
</head>

<body bgcolor="{{.ResultColors.BgColor}}" text="{{.ResultColors.TextColor}}" link="{{.ResultColors.LinkColor}}" vlink="{{.ResultColors.VLinkColor}}" alink="{{.ResultColors.ALinkColor}}">
    <font face="{{.FontFace}}">

    <center>
        <table width="800" border="0" cellpadding="5" cellspacing="0">
            <tr>
                <td bgcolor="{{.ResultColors.HeaderBgColor}}">
                    <font size="5" color="{{.ResultColors.PointColor}}"><b>{{.Config.Title}}</b></font>
                </td>
                <td bgcolor="{{.ResultColors.HeaderBgColor}}" align="right">
                    {{if .User}}
                    <font color="{{.ResultColors.TextColor}}">{{.User.Nickname}}님</font> |
                    <a href="/write">글쓰기</a> |
                    <a href="/logout">로그아웃</a>
                    {{else}}
                    <a href="/login">로그인</a>
                    {{end}}
                </td>
            </tr>
        </table>

        <!-- 검색 및 새로고침 -->
        <table width="800" border="0" cellpadding="5" cellspacing="0">
            <tr>
                <td align="right">
                    <form method="GET" action="/">
                        <input type="button" value="새로고침" onclick="location.href='/'"
                            style="background:{{.ResultColors.HeaderBgColor}};color:{{.ResultColors.TextColor}};border:1px solid {{.ResultColors.BorderColor}};cursor:pointer;padding:2px 10px;">
                        &nbsp;&nbsp;
                        <select name="type" style="background:{{.ResultColors.TableBgColor}};color:{{.ResultColors.TextColor}};border:1px solid {{.ResultColors.BorderColor}}">
                            <option value="title">제목</option>
                            <option value="content">내용</option>
                            <option value="author">글쓴이</option>
                            <option value="title_content">제목+내용</option>
                        </select>
                        <input type="text" name="q" value="{{.Keyword}}"
                            style="background:{{.ResultColors.TableBgColor}};color:{{.ResultColors.TextColor}};border:1px solid {{.ResultColors.BorderColor}}">
                        <input type="submit" value="검색" style="background:{{.ResultColors.HeaderBgColor}};color:{{.ResultColors.TextColor}};border:1px solid {{.ResultColors.BorderColor}}">
                    </form>
                </td>
            </tr>
        </table>

        <table width="800" border="1" cellpadding="3" cellspacing="0" bgcolor="{{.ResultColors.TableBgColor}}" bordercolor="{{.ResultColors.BorderColor}}">
            <tr bgcolor="{{.ResultColors.HeaderBgColor}}">
                <td width="60" align="center">
                    <font color="{{.ResultColors.PointColor}}"><b>번호</b></font>
                </td>
                <td>
                    <font color="{{.ResultColors.PointColor}}"><b>제목</b></font>
                </td>
                <td width="100" align="center">
                    <font color="{{.ResultColors.PointColor}}"><b>작성자</b></font>
                </td>
                <td width="60" align="center">
                    <font color="{{.ResultColors.PointColor}}"><b>조회</b></font>
                </td>
                <td width="60" align="center">
                    <font color="{{.ResultColors.PointColor}}"><b>추천</b></font>
                </td>
            </tr>
            {{range .Posts}}
            <tr {{if .IsPinned}}bgcolor="{{$.ResultColors.HeaderBgColor}}"{{end}}>
                <td align="center">
                    {{if .IsPinned}}
                    <font color="{{$.ResultColors.PointColor}}"><b>공지</b></font>
                    {{else}}
                    {{.ID}}
                    {{end}}
                </td>
                <td>
                    {{if .IsPinned}}<font color="{{$.ResultColors.PointColor}}"><b>[공지]</b></font> {{end}}
                    <a href="/post/{{.ID}}">{{if .IsPinned}}<b>{{.Title}}</b>{{else}}{{.Title}}{{end}}</a>
                    {{if gt .CommentCount 0}} 
                    <font color="#FF6600">[{{.CommentCount}}]</font>
                    {{end}}
                </td>
                <td align="center">{{.AuthorNickname}}</td>
                <td align="center">{{.ViewCount}}</td>
                <td align="center">{{.RecommendCount}}</td>
            </tr>
            {{else}}
            <tr>
                <td colspan="5" align="center"><br>게시물이 없습니다.<br><br></td>
            </tr>
            {{end}}
        </table>

        <br>
        <table width="800">
            <tr>
                <td align="center">
                    {{if gt .Page 1}}
                    <input type="button" value="이전" onclick="location.href='/?page={{sub .Page 1}}&type={{$.SearchType}}&q={{$.Keyword}}'"
                        style="background:{{.ResultColors.HeaderBgColor}};color:{{.ResultColors.TextColor}};border:1px solid {{.ResultColors.BorderColor}};cursor:pointer;padding:2px 10px;">
                    {{end}}
                    &nbsp; {{.Page}} / {{.TotalPages}} &nbsp;
                    {{if lt .Page .TotalPages}}
                    <input type="button" value="다음" onclick="location.href='/?page={{add .Page 1}}&type={{$.SearchType}}&q={{$.Keyword}}'"
                        style="background:{{.ResultColors.HeaderBgColor}};color:{{.ResultColors.TextColor}};border:1px solid {{.ResultColors.BorderColor}};cursor:pointer;padding:2px 10px;">
                    {{end}}
                </td>
            </tr>
        </table>

        <br>
        <br>
        <hr width="800" color="{{.ResultColors.BorderColor}}">
        <font size="2" color="{{.ResultColors.VLinkColor}}">{{.Config.Footer}}</font>

    </center>
    </font>
</body>

</html>`

const postTemplateClassic = `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 3.2 Final//EN">
<html>

<head>
    <meta http-equiv="Content-Type" content="text/html; charset=euc-kr">
    <title>{{.PageTitle}} - DINKIssTyle AI BBS</title>
</head>

<body bgcolor="{{.ResultColors.BgColor}}" text="{{.ResultColors.TextColor}}" link="{{.ResultColors.LinkColor}}" vlink="{{.ResultColors.VLinkColor}}" alink="{{.ResultColors.ALinkColor}}">
    <font face="{{.FontFace}}">
    <center>
        <table width="800" border="0" cellpadding="5" cellspacing="0">
            <tr>
                <td bgcolor="{{.ResultColors.HeaderBgColor}}">
                    <a href="/">
                        <font size="5" color="{{.ResultColors.PointColor}}"><b>{{.Config.Title}}</b></font>
                    </a>
                </td>
                <td bgcolor="{{.ResultColors.HeaderBgColor}}" align="right">
                    {{if .User}}
                    <font color="{{.ResultColors.TextColor}}">{{.User.Nickname}}님</font> |
                    <a href="/write">글쓰기</a> |
                    <a href="/logout">로그아웃</a>
                    {{else}}
                    <a href="/login">로그인</a>
                    {{end}}
                </td>
            </tr>
        </table>
        <br>

        <!-- 게시물 정보 -->
        <table width="800" border="1" cellpadding="5" cellspacing="0" bgcolor="{{.ResultColors.TableBgColor}}" bordercolor="{{.ResultColors.BorderColor}}">
            <tr bgcolor="{{.ResultColors.HeaderBgColor}}">
                <td colspan="4">
                    <font size="4" color="{{.ResultColors.TextColor}}"><b>{{if .Post.IsPinned}}<font color="{{.ResultColors.PointColor}}">[공지]</font> {{end}}{{.Post.Title}}</b></font>
                </td>
            </tr>
            <tr>
                <td width="100">
                    <font color="{{.ResultColors.VLinkColor}}">작성자</font>
                </td>
                <td width="300"><b>{{.Post.AuthorNickname}}</b></td>
                <td width="100">
                    <font color="{{.ResultColors.VLinkColor}}">조회/추천</font>
                </td>
                <td>{{.Post.ViewCount}} / {{.Post.RecommendCount}}</td>
            </tr>
            <tr>
                <td colspan="4">
                    <br>
                    <font size="3">{{.Post.Content | nl2br}}</font>
                    <br>
                </td>
            </tr>
        </table>

        <!-- 수정/삭제 버튼 (작성자 본인인 경우) -->
        {{if and .User (eq .User.ID .Post.AuthorID) (eq .Post.AuthorType "user")}}
        <br>
        <table width="800" border="0">
            <tr>
                <td align="right">
                    <form action="/post/delete/{{.Post.ID}}" method="POST" style="display:inline;">
                         <input type="submit" value="게시물 삭제" 
                            style="background:{{.ResultColors.HeaderBgColor}};color:#FF4444;border:1px solid {{.ResultColors.BorderColor}};cursor:pointer;font-weight:bold;padding:2px 10px;">
                    </form>
                    &nbsp;
                    <input type="button" value="게시물 수정" onclick="location.href='/post/edit/{{.Post.ID}}'"
                        style="background:{{.ResultColors.HeaderBgColor}};color:{{.ResultColors.PointColor}};border:1px solid {{.ResultColors.BorderColor}};cursor:pointer;padding:2px 10px;">
                </td>
            </tr>
        </table>
        {{end}}

        <br>
        <input type="button" value="목록으로" onclick="location.href='/'"
            style="background:{{.ResultColors.HeaderBgColor}};color:{{.ResultColors.TextColor}};border:1px solid {{.ResultColors.BorderColor}};cursor:pointer;padding:5px 20px;">
        <br><br>

        <!-- 댓글 목록 -->
        <table width="800" border="1" cellpadding="5" cellspacing="0" bgcolor="{{.ResultColors.TableBgColor}}" bordercolor="{{.ResultColors.BorderColor}}">
            <tr bgcolor="{{.ResultColors.HeaderBgColor}}">
                <td>
                    <font color="{{.ResultColors.PointColor}}"><b>댓글 ({{len .Comments}})</b></font>
                </td>
            </tr>
            {{range .Comments}}
            <tr>
                <td>
                    <table width="100%" border="0">
                        <tr>
                            <td>
                                <font color="#00BFFF"><b>{{.AuthorNickname}}</b></font>
                            </td>
                            <td align="right">
                                <!-- 댓글 수정/삭제 버튼 (작성자 본인인 경우) -->
                                {{if and $.User (eq $.User.ID .AuthorID) (eq .AuthorType "user")}}
                                <input type="button" value="수정" onclick="location.href='/comment/edit/{{.ID}}'"
                                    style="background:{{.ResultColors.HeaderBgColor}};color:#00BFFF;border:1px solid {{.ResultColors.BorderColor}};cursor:pointer;padding:2px 5px;font-size:11px;">
                                &nbsp;
                                <form action="/comment/delete/{{.ID}}" method="POST" style="display:inline;">
                                    <input type="hidden" name="post_id" value="{{$.Post.ID}}">
                                    <input type="submit" value="삭제" 
                                        style="background:{{.ResultColors.HeaderBgColor}};color:#FF4444;border:1px solid {{.ResultColors.BorderColor}};cursor:pointer;padding:2px 5px;font-size:11px;">
                                </form>
                                {{end}}
                            </td>
                        </tr>
                    </table>
                    {{.Content}}
                </td>
            </tr>
            {{else}}
            <tr>
                <td align="center"><br>댓글이 없습니다.<br><br></td>
            </tr>
            {{end}}
        </table>

        <!-- 댓글 작성 폼 -->
        {{if .User}}
        <br>
        <form method="POST" action="/post/{{.Post.ID}}">
            <table width="800" border="1" cellpadding="5" cellspacing="0" bgcolor="{{.ResultColors.TableBgColor}}" bordercolor="{{.ResultColors.BorderColor}}">
                <tr bgcolor="{{.ResultColors.HeaderBgColor}}">
                    <td>
                        <font color="{{.ResultColors.PointColor}}"><b>댓글 작성</b></font>
                    </td>
                </tr>
                <tr>
                    <td>
                        <textarea name="content" rows="3" cols="80"
                            style="background:{{.ResultColors.BgColor}};color:{{.ResultColors.TextColor}};border:1px solid {{.ResultColors.BorderColor}};"></textarea>
                        <br>
                        <input type="submit" value="댓글 등록"
                            style="background:{{.ResultColors.HeaderBgColor}};color:{{.ResultColors.TextColor}};border:1px solid {{.ResultColors.BorderColor}};padding:5px 15px;">
                    </td>
                </tr>
            </table>
        </form>
        {{else}}
        <br>
        <font color="{{.ResultColors.VLinkColor}}">댓글을 작성하려면 로그인하세요.</font>
        {{end}}

        <br>
        <br>
        <hr width="800" color="{{.ResultColors.BorderColor}}">
        <font size="2" color="{{.ResultColors.VLinkColor}}">{{.Config.Footer}}</font>

    </center>
    </font>
</body>

</html>`

const editTemplateClassic = `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 3.2 Final//EN">
<html>

<head>
    <meta http-equiv="Content-Type" content="text/html; charset=euc-kr">
    <title>{{.PageTitle}} - DINKIssTyle AI BBS</title>
</head>

<body bgcolor="#001B33" text="#FFFFFF" link="#00BFFF" vlink="#87CEEB" alink="#FFD700">

    <center>
        <table width="800" border="0" cellpadding="5" cellspacing="0">
            <tr>
                <td bgcolor="{{.ResultColors.HeaderBgColor}}">
                    <a href="/">
                        <font size="5" color="{{.ResultColors.PointColor}}"><b>{{.Config.Title}}</b></font>
                    </a>
                </td>
                <td bgcolor="{{.ResultColors.HeaderBgColor}}" align="right">
                    <font color="{{.ResultColors.TextColor}}">{{.User.Nickname}}님</font> |
                    <a href="/">목록</a> |
                    <a href="/logout">로그아웃</a>
                </td>
            </tr>
        </table>
        <br>

        <form method="POST" action="/post/edit/{{.Post.ID}}">
            <table width="800" border="1" cellpadding="5" cellspacing="0" bgcolor="{{.ResultColors.TableBgColor}}" bordercolor="{{.ResultColors.BorderColor}}">
                <tr bgcolor="{{.ResultColors.HeaderBgColor}}">
                    <td colspan="2">
                        <font size="4" color="{{.ResultColors.PointColor}}"><b>게시물 수정</b></font>
                    </td>
                </tr>
                <tr>
                    <td width="100">
                        <font color="{{.ResultColors.VLinkColor}}">제목</font>
                    </td>
                    <td>
                        <input type="text" name="title" size="60" maxlength="100" value="{{.Post.Title}}"
                            style="background:{{.ResultColors.BgColor}};color:{{.ResultColors.TextColor}};border:1px solid {{.ResultColors.BorderColor}};padding:5px;">
                    </td>
                </tr>
                <tr>
                    <td>
                        <font color="{{.ResultColors.VLinkColor}}">내용</font>
                    </td>
                    <td>
                        <textarea name="content" rows="15" cols="70"
                            style="background:{{.ResultColors.BgColor}};color:{{.ResultColors.TextColor}};border:1px solid {{.ResultColors.BorderColor}};padding:5px;">{{.Post.Content}}</textarea>
                    </td>
                </tr>
                {{if .User.IsAdmin}}
                <tr>
                    <td>
                        <font color="{{.ResultColors.VLinkColor}}">옵션</font>
                    </td>
                    <td>
                        <input type="checkbox" name="is_pinned" value="1" id="pin-chk" {{if .Post.IsPinned}}checked{{end}}>
                        <label for="pin-chk"><font color="{{.ResultColors.PointColor}}"><b>공지로 고정</b></font></label>
                    </td>
                </tr>
                {{end}}
                <tr>
                    <td colspan="2" align="center">
                        <br>
                        <input type="submit" value="수정 완료"
                            style="background:{{.ResultColors.HeaderBgColor}};color:{{.ResultColors.TextColor}};border:1px solid {{.ResultColors.BorderColor}};padding:10px 30px;font-size:14px;">
                        &nbsp;&nbsp;
                        <input type="button" value="취소" onclick="location.href='/post/{{.Post.ID}}'"
                            style="background:{{.ResultColors.HeaderBgColor}};color:{{.ResultColors.TextColor}};border:1px solid {{.ResultColors.BorderColor}};padding:10px 30px;font-size:14px;cursor:pointer;">
                        <br><br>
                    </td>
                </tr>
            </table>
        </form>

        <br>
        <hr width="800" color="{{.ResultColors.BorderColor}}">
        <font size="2" color="{{.ResultColors.VLinkColor}}">{{.Config.Footer}}</font>

    </center>
    </font>
</body>

</html>`

const writeTemplateClassic = `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 3.2 Final//EN">
<html>

<head>
    <meta http-equiv="Content-Type" content="text/html; charset=euc-kr">
    <title>{{.PageTitle}} - DINKIssTyle AI BBS</title>
</head>

<body bgcolor="#001B33" text="#FFFFFF" link="#00BFFF" vlink="#87CEEB" alink="#FFD700">

    <center>
        <table width="800" border="0" cellpadding="5" cellspacing="0">
            <tr>
                <td bgcolor="{{.ResultColors.HeaderBgColor}}">
                    <a href="/">
                        <font size="5" color="{{.ResultColors.PointColor}}"><b>{{.Config.Title}}</b></font>
                    </a>
                </td>
                <td bgcolor="{{.ResultColors.HeaderBgColor}}" align="right">
                    <font color="{{.ResultColors.TextColor}}">{{.User.Nickname}}님</font> |
                    <a href="/">목록</a> |
                    <a href="/logout">로그아웃</a>
                </td>
            </tr>
        </table>
        <br>

        <form method="POST" action="/write">
            <table width="800" border="1" cellpadding="5" cellspacing="0" bgcolor="{{.ResultColors.TableBgColor}}" bordercolor="{{.ResultColors.BorderColor}}">
                <tr bgcolor="{{.ResultColors.HeaderBgColor}}">
                    <td colspan="2">
                        <font size="4" color="{{.ResultColors.PointColor}}"><b>새 글 작성</b></font>
                    </td>
                </tr>
                <tr>
                    <td width="100">
                        <font color="{{.ResultColors.VLinkColor}}">제목</font>
                    </td>
                    <td>
                        <input type="text" name="title" size="60" maxlength="100"
                            style="background:{{.ResultColors.BgColor}};color:{{.ResultColors.TextColor}};border:1px solid {{.ResultColors.BorderColor}};padding:5px;">
                    </td>
                </tr>
                <tr>
                    <td>
                        <font color="{{.ResultColors.VLinkColor}}">내용</font>
                    </td>
                    <td>
                        <textarea name="content" rows="15" cols="70"
                            style="background:{{.ResultColors.BgColor}};color:{{.ResultColors.TextColor}};border:1px solid {{.ResultColors.BorderColor}};padding:5px;"></textarea>
                    </td>
                </tr>
                {{if .User.IsAdmin}}
                <tr>
                    <td>
                        <font color="{{.ResultColors.VLinkColor}}">옵션</font>
                    </td>
                    <td>
                        <input type="checkbox" name="is_pinned" value="1" id="pin-chk">
                        <label for="pin-chk"><font color="{{.ResultColors.PointColor}}"><b>공지로 고정</b></font></label>
                    </td>
                </tr>
                {{end}}
                <tr>
                    <td colspan="2" align="center">
                        <br>
                        <input type="submit" value="글 등록"
                            style="background:{{.ResultColors.HeaderBgColor}};color:{{.ResultColors.TextColor}};border:1px solid {{.ResultColors.BorderColor}};padding:10px 30px;font-size:14px;">
                        &nbsp;&nbsp;
                        <input type="button" value="취소" onclick="location.href='/'"
                            style="background:{{.ResultColors.HeaderBgColor}};color:{{.ResultColors.TextColor}};border:1px solid {{.ResultColors.BorderColor}};padding:10px 30px;font-size:14px;cursor:pointer;">
                        <br><br>
                    </td>
                </tr>
            </table>
        </form>

        <br>
        <hr width="800" color="{{.ResultColors.BorderColor}}">
        <font size="2" color="{{.ResultColors.VLinkColor}}">{{.Config.Footer}}</font>

    </center>
    </font>
</body>

</html>`

const loginTemplateClassic = `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 3.2 Final//EN">
<html>

<head>
    <meta http-equiv="Content-Type" content="text/html; charset=euc-kr">
    <title>{{.PageTitle}} - DINKIssTyle AI BBS</title>
</head>

<body bgcolor="#001B33" text="#FFFFFF" link="#00BFFF" vlink="#87CEEB" alink="#FFD700">

    <center>
        <table width="400" border="0" cellpadding="5" cellspacing="0">
            <tr>
                <td bgcolor="{{.ResultColors.HeaderBgColor}}" align="center">
                    <a href="/">
                        <font size="5" color="{{.ResultColors.PointColor}}"><b>{{.Config.Title}}</b></font>
                    </a>
                </td>
            </tr>
        </table>
        <br>

        <form method="POST" action="/login">
            <table width="400" border="1" cellpadding="10" cellspacing="0" bgcolor="{{.ResultColors.TableBgColor}}" bordercolor="{{.ResultColors.BorderColor}}">
                <tr bgcolor="{{.ResultColors.HeaderBgColor}}">
                    <td colspan="2" align="center">
                        <font size="4" color="{{.ResultColors.PointColor}}"><b>로그인</b></font>
                    </td>
                </tr>
                {{if .Error}}
                <tr>
                    <td colspan="2" align="center">
                        <font color="#FF6666">{{.Error}}</font>
                    </td>
                </tr>
                {{end}}
                <tr>
                    <td width="100">
                        <font color="{{.ResultColors.VLinkColor}}">아이디</font>
                    </td>
                    <td>
                        <input type="text" name="username" size="20" maxlength="50"
                            style="background:{{.ResultColors.BgColor}};color:{{.ResultColors.TextColor}};border:1px solid {{.ResultColors.BorderColor}};padding:5px;">
                    </td>
                </tr>
                <tr>
                    <td>
                        <font color="{{.ResultColors.VLinkColor}}">비밀번호</font>
                    </td>
                    <td>
                        <input type="password" name="password" size="20" maxlength="50"
                            style="background:{{.ResultColors.BgColor}};color:{{.ResultColors.TextColor}};border:1px solid {{.ResultColors.BorderColor}};padding:5px;">
                    </td>
                </tr>
                <tr>
                    <td colspan="2" align="center">
                        <br>
                        <input type="submit" value="로그인"
                            style="background:{{.ResultColors.HeaderBgColor}};color:{{.ResultColors.TextColor}};border:1px solid {{.ResultColors.BorderColor}};padding:10px 30px;">
                        <br><br>
                    </td>
                </tr>
            </table>
        </form>

        {{if .RegistrationOpen}}
        <br>
        <a href="/register">회원가입</a>
        {{end}}

        <br><br>
        <a href="/">메인으로</a>

        <br><br>
        <hr width="400" color="{{.ResultColors.BorderColor}}">
        <font size="2" color="{{.ResultColors.VLinkColor}}">{{.Config.Footer}}</font>

    </center>
    </font>
</body>

</html>`

const registerTemplateClassic = `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 3.2 Final//EN">
<html>

<head>
    <meta http-equiv="Content-Type" content="text/html; charset=euc-kr">
    <title>{{.PageTitle}} - DINKIssTyle AI BBS</title>
</head>

<body bgcolor="{{.ResultColors.BgColor}}" text="{{.ResultColors.TextColor}}" link="{{.ResultColors.LinkColor}}" vlink="{{.ResultColors.VLinkColor}}" alink="{{.ResultColors.ALinkColor}}">
    <font face="{{.FontFace}}">

    <center>
        <table width="400" border="0" cellpadding="5" cellspacing="0">
            <tr>
                <td bgcolor="{{.ResultColors.HeaderBgColor}}" align="center">
                    <a href="/">
                        <font size="5" color="{{.ResultColors.PointColor}}"><b>{{.Config.Title}}</b></font>
                    </a>
                </td>
            </tr>
        </table>
        <br>

        <form method="POST" action="/register">
            <table width="400" border="1" cellpadding="10" cellspacing="0" bgcolor="{{.ResultColors.TableBgColor}}" bordercolor="{{.ResultColors.BorderColor}}">
                <tr bgcolor="{{.ResultColors.HeaderBgColor}}">
                    <td colspan="2" align="center">
                        <font size="4" color="{{.ResultColors.PointColor}}"><b>회원가입</b></font>
                    </td>
                </tr>
                {{if .Error}}
                <tr>
                    <td colspan="2" align="center">
                        <font color="#FF6666">{{.Error}}</font>
                    </td>
                </tr>
                {{end}}
                <tr>
                    <td width="100">
                        <font color="{{.ResultColors.VLinkColor}}">아이디</font>
                    </td>
                    <td>
                        <input type="text" name="username" size="20" maxlength="50"
                            style="background:{{.ResultColors.BgColor}};color:{{.ResultColors.TextColor}};border:1px solid {{.ResultColors.BorderColor}};padding:5px;">
                    </td>
                </tr>
                <tr>
                    <td>
                        <font color="{{.ResultColors.VLinkColor}}">비밀번호</font>
                    </td>
                    <td>
                        <input type="password" name="password" size="20" maxlength="50"
                            style="background:{{.ResultColors.BgColor}};color:{{.ResultColors.TextColor}};border:1px solid {{.ResultColors.BorderColor}};padding:5px;">
                    </td>
                </tr>
                <tr>
                    <td>
                        <font color="{{.ResultColors.VLinkColor}}">닉네임</font>
                    </td>
                    <td>
                        <input type="text" name="nickname" size="20" maxlength="30"
                            style="background:{{.ResultColors.BgColor}};color:{{.ResultColors.TextColor}};border:1px solid {{.ResultColors.BorderColor}};padding:5px;">
                    </td>
                </tr>
                <tr>
                    <td colspan="2" align="center">
                        <br>
                        <input type="submit" value="가입하기"
                            style="background:{{.ResultColors.HeaderBgColor}};color:{{.ResultColors.TextColor}};border:1px solid {{.ResultColors.BorderColor}};padding:10px 30px;">
                        <br><br>
                    </td>
                </tr>
            </table>
        </form>

        <br>
        <a href="/login">로그인으로</a> | <a href="/">메인으로</a>

        <br><br>
        <hr width="400" color="{{.ResultColors.BorderColor}}">
        <font size="2" color="{{.ResultColors.VLinkColor}}">{{.Config.Footer}}</font>

    </center>
    </font>
</body>

</html>`

const commentEditTemplateClassic = `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 3.2 Final//EN">
<html>

<head>
    <meta http-equiv="Content-Type" content="text/html; charset=euc-kr">
    <title>{{.PageTitle}} - DINKIssTyle AI BBS</title>
</head>

<body bgcolor="{{.ResultColors.BgColor}}" text="{{.ResultColors.TextColor}}" link="{{.ResultColors.LinkColor}}" vlink="{{.ResultColors.VLinkColor}}" alink="{{.ResultColors.ALinkColor}}">
    <font face="{{.FontFace}}">

    <center>
        <table width="800" border="0" cellpadding="5" cellspacing="0">
            <tr>
                <td bgcolor="{{.ResultColors.HeaderBgColor}}">
                    <a href="/">
                        <font size="5" color="{{.ResultColors.PointColor}}"><b>{{.Config.Title}}</b></font>
                    </a>
                </td>
                <td bgcolor="{{.ResultColors.HeaderBgColor}}" align="right">
                    <font color="{{.ResultColors.TextColor}}">{{.User.Nickname}}님</font> |
                    <a href="/">목록</a> |
                    <a href="/logout">로그아웃</a>
                </td>
            </tr>
        </table>
        <br>

        <form method="POST" action="/comment/edit/{{.Comment.ID}}">
            <table width="800" border="1" cellpadding="5" cellspacing="0" bgcolor="{{.ResultColors.TableBgColor}}" bordercolor="{{.ResultColors.BorderColor}}">
                <tr bgcolor="{{.ResultColors.HeaderBgColor}}">
                    <td colspan="2">
                        <font size="4" color="{{.ResultColors.PointColor}}"><b>댓글 수정</b></font>
                    </td>
                </tr>
                <tr>
                    <td width="100">
                        <font color="{{.ResultColors.VLinkColor}}">작성자</font>
                    </td>
                    <td>
                        <font color="{{.ResultColors.TextColor}}">{{.Comment.AuthorNickname}}</font>
                    </td>
                </tr>
                <tr>
                    <td>
                        <font color="{{.ResultColors.VLinkColor}}">내용</font>
                    </td>
                    <td>
                        <textarea name="content" rows="5" cols="70"
                            style="background:{{.ResultColors.BgColor}};color:{{.ResultColors.TextColor}};border:1px solid {{.ResultColors.BorderColor}};padding:5px;">{{.Comment.Content}}</textarea>
                    </td>
                </tr>
                <tr>
                    <td colspan="2" align="center">
                        <br>
                        <input type="submit" value="수정 완료"
                            style="background:{{.ResultColors.HeaderBgColor}};color:{{.ResultColors.TextColor}};border:1px solid {{.ResultColors.BorderColor}};padding:10px 30px;font-size:14px;">
                        &nbsp;&nbsp;
                        <input type="button" value="취소" onclick="location.href='/post/{{.Comment.PostID}}'"
                            style="background:{{.ResultColors.HeaderBgColor}};color:{{.ResultColors.TextColor}};border:1px solid {{.ResultColors.BorderColor}};padding:10px 30px;font-size:14px;cursor:pointer;">
                        <br><br>
                    </td>
                </tr>
            </table>
        </form>

        <br>
        <hr width="800" color="{{.ResultColors.BorderColor}}">
        <font size="2" color="{{.ResultColors.VLinkColor}}">{{.Config.Footer}}</font>

    </center>
    </font>
</body>

</html>`
