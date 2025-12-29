// Created by DINKIssTyle on 2025. Copyright (C) 2025 DINKI'ssTyle. All rights reserved.

package web

// Mobile 전용 미니멀 템플릿

const boardTemplateMobile = `<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
    <title>{{.PageTitle}}</title>
    <style>
        body { margin: 0; padding: 0; background: {{.ResultColors.BgColor}}; color: {{.ResultColors.TextColor}}; font-family: sans-serif; }
        .header { background: {{.ResultColors.HeaderBgColor}}; padding: 15px; display: flex; justify-content: space-between; align-items: center; }
        .header h1 { margin: 0; font-size: 1.2rem; color: {{.ResultColors.PointColor}}; }
        .post-list { list-style: none; margin: 0; padding: 0; }
        .post-item { border-bottom: 1px solid {{.ResultColors.BorderColor}}; padding: 15px; }
        .post-item a { text-decoration: none; color: {{.ResultColors.TextColor}}; font-weight: bold; }
        .post-meta { font-size: 0.8rem; color: #888; margin-top: 5px; }
        .btn { display: inline-block; padding: 10px 15px; background: {{.ResultColors.PointColor}}; color: #fff; text-decoration: none; border-radius: 5px; border: none; font-size: 0.9rem; }
        .search-area { padding: 15px; background: rgba(0,0,0,0.2); }
        .search-area input { width: 100%; box-sizing: border-box; padding: 10px; margin-bottom: 5px; background: #000; color: #fff; border: 1px solid #444; }
        .pagination { padding: 20px; text-align: center; }
    </style>
</head>
<body>
    <div class="header">
        <h1>{{.Config.Title}}</h1>
        <div>
            {{if .User}}<a href="/write" class="btn">글쓰기</a>{{else}}<a href="/login" style="color:#fff;">로그인</a>{{end}}
        </div>
    </div>
    
    <ul class="post-list">
        {{range .Posts}}
        <li class="post-item" {{if .IsPinned}}style="background:rgba(255,215,0,0.1); border-left: 4px solid {{$.ResultColors.PointColor}};"{{end}}>
            {{if .IsPinned}}<span style="color:{{$.ResultColors.PointColor}}; font-weight:bold;">[공지]</span> {{end}}
            <a href="/post/{{.ID}}">{{if .IsPinned}}<b>{{.Title}}</b>{{else}}{{.Title}}{{end}}</a>
            {{if gt .CommentCount 0}}<span style="color:#f60;">[{{.CommentCount}}]</span>{{end}}
            <div class="post-meta">{{.AuthorNickname}} | 조회 {{.ViewCount}}</div>
        </li>
        {{else}}
        <li class="post-item" style="text-align:center;">게시물이 없습니다.</li>
        {{end}}
    </ul>

    <div class="pagination">
        {{if gt .Page 1}}<a href="/?page={{sub .Page 1}}" class="btn">이전</a>{{end}}
        <span>{{.Page}} / {{.TotalPages}}</span>
        {{if lt .Page .TotalPages}}<a href="/?page={{add .Page 1}}" class="btn">다음</a>{{end}}
    </div>
</body>
</html>`

const postTemplateMobile = `<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
    <title>{{.Post.Title}}</title>
    <style>
        body { margin: 0; padding: 0; background: {{.ResultColors.BgColor}}; color: {{.ResultColors.TextColor}}; font-family: sans-serif; }
        .header { background: {{.ResultColors.HeaderBgColor}}; padding: 15px; }
        .container { padding: 20px; }
        .meta { color: #888; font-size: 0.9rem; margin-bottom: 20px; border-bottom: 1px solid {{.ResultColors.BorderColor}}; padding-bottom: 10px; }
        .content { line-height: 1.6; font-size: 1rem; margin-bottom: 30px; border-bottom: 1px solid {{.ResultColors.BorderColor}}; padding-bottom: 30px; white-space: pre-wrap; }
        .btn { display: inline-block; padding: 10px 15px; background: {{.ResultColors.PointColor}}; color: #fff; text-decoration: none; border: none; border-radius: 5px; cursor: pointer; }
        .comment { background: rgba(255,255,255,0.05); padding: 15px; margin-bottom: 10px; border-radius: 5px; }
        .comment-header { font-weight: bold; margin-bottom: 5px; color: {{.ResultColors.PointColor}}; font-size: 0.9rem; }
    </style>
</head>
<body>
    <div class="header">
        <a href="/" style="color:{{.ResultColors.PointColor}}; text-decoration:none;">&larr; {{.Config.Title}}</a>
    </div>
    <div class="container">
        <h2 style="margin-top:0;">{{if .Post.IsPinned}}<span style="color:{{.ResultColors.PointColor}};">[공지]</span> {{end}}{{.Post.Title}}</h2>
        <div class="meta">{{.Post.AuthorNickname}} | {{.Post.CreatedAt | formatDate}} | 조회 {{.Post.ViewCount}} | 추천 {{.Post.RecommendCount}}</div>
        <div class="content">{{.Post.Content}}</div>

        <div style="margin-bottom:40px; display:flex; gap:10px;">
            <form action="/post/recommend/{{.Post.ID}}" method="POST" style="margin:0;">
                <input type="submit" value="👍 추천" class="btn" style="background:#444;">
            </form>
            <a href="/" class="btn" style="background:#555;">목록</a>
            {{if and .User (eq .User.ID .Post.AuthorID) (eq .Post.AuthorType "user")}}
            <a href="/post/edit/{{.Post.ID}}" class="btn">수정</a>
            {{end}}
        </div>

        <h3>댓글 ({{len .Comments}})</h3>
        {{range .Comments}}
        <div class="comment">
            <div class="comment-header">{{.AuthorNickname}} <span style="color:#666; font-weight:normal; font-size:0.7rem;">{{.CreatedAt | formatDate}}</span></div>
            <div>{{.Content}}</div>
        </div>
        {{end}}

        {{if .User}}
        <form method="POST" action="/post/{{.Post.ID}}" style="margin-top:20px;">
            <textarea name="content" style="width:100%; background:#000; color:#fff; border:1px solid #444; padding:10px; box-sizing:border-box;" rows="3"></textarea>
            <input type="submit" value="댓글 등록" class="btn" style="width:100%; margin-top:5px;">
        </form>
        {{end}}
    </div>
</body>
</html>`

const loginTemplateMobile = `<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
    <title>로그인</title>
    <style>
        body { background: {{.ResultColors.BgColor}}; color: #fff; font-family: sans-serif; display: flex; align-items: center; justify-content: center; height: 100vh; margin: 0; }
        .card { width: 85%; background: rgba(255,255,255,0.05); padding: 25px; border-radius: 15px; border: 1px solid #333; }
        input { width: 100%; padding: 12px; margin-bottom: 10px; box-sizing: border-box; border-radius: 5px; border: 1px solid #444; background: #000; color: #fff; }
    </style>
</head>
<body>
    <div class="card">
        <h2 align="center" style="color:{{.ResultColors.PointColor}};">{{.Config.Title}}</h2>
        {{if .Error}}<p style="color:#f44; font-size:0.9rem;">{{.Error}}</p>{{end}}
        <form method="POST" action="/login">
            <input type="text" name="username" placeholder="아이디">
            <input type="password" name="password" placeholder="비밀번호">
            <input type="submit" value="로그인" style="background:{{.ResultColors.PointColor}}; border:none; font-weight:bold;">
        </form>
        <div align="center" style="margin-top:15px; font-size:0.9rem;"><a href="/" style="color:#888;">메인으로</a> | <a href="/register" style="color:#888;">회원가입</a></div>
    </div>
</body>
</html>`

const writeTemplateMobile = `<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
    <title>글쓰기</title>
    <style>
        body { background: {{.ResultColors.BgColor}}; color: #fff; font-family: sans-serif; margin: 0; }
        .header { background: {{.ResultColors.HeaderBgColor}}; padding: 15px; }
        .container { padding: 15px; }
        input, textarea { width: 100%; border-radius: 5px; border: 1px solid #444; background: #000; color: #fff; padding: 12px; box-sizing: border-box; margin-bottom: 15px; }
        .btn { background: {{.ResultColors.PointColor}}; color: #fff; border: none; padding: 12px; width: 100%; font-weight: bold; border-radius: 5px; }
    </style>
</head>
<body>
    <div class="header">
        <a href="/" style="color:{{.ResultColors.PointColor}}; text-decoration:none;">&larr; 돌아가기</a>
    </div>
    <div class="container">
        <h3>새 글 작성</h3>
        <form method="POST" action="/write">
            <input type="text" name="title" placeholder="제목">
            <textarea name="content" rows="10" placeholder="내용"></textarea>
            {{if .User.IsAdmin}}
            <div style="margin-bottom: 15px;">
                <label style="display: flex; align-items: center; gap: 10px;">
                    <input type="checkbox" name="is_pinned" value="1" style="width: auto; margin: 0;"> 
                    <span style="color: {{.ResultColors.PointColor}}; font-weight: bold;">공지로 고정</span>
                </label>
            </div>
            {{end}}
            <input type="submit" value="등록하기" class="btn">
        </form>
    </div>
</body>
</html>`
