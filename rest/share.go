package rest

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/saltbo/gopkg/ginutil"
	"github.com/saltbo/gopkg/jwtutil"
	"github.com/saltbo/gopkg/strutil"
	"gorm.io/gorm"

	"github.com/saltbo/zpan/dao"
	"github.com/saltbo/zpan/dao/matter"
	"github.com/saltbo/zpan/model"
	"github.com/saltbo/zpan/pkg/gormutil"
	"github.com/saltbo/zpan/rest/bind"
)

const ShareCookieTokenKey = "share-token"

type ShareResource struct {
	jwtutil.JWTUtil

	dShare  *dao.Share
	dMatter *matter.Matter
}

func NewShareResource() ginutil.Resource {
	return &ShareResource{
		dShare:  dao.NewShare(),
		dMatter: matter.NewMatter(),
	}
}

func (rs *ShareResource) Register(router *gin.RouterGroup) {
	router.GET("/shares/:alias", rs.find)
	router.GET("/shares", rs.findAll)
	router.POST("/shares", rs.create)
	router.PATCH("/shares/:alias", rs.update)
	router.DELETE("/shares/:alias", rs.delete)

	router.POST("/shares/:alias/token", rs.draw)
	router.GET("/shares/:alias/matter", rs.findMatter)
	router.GET("/shares/:alias/matters", rs.findMatters)
}

func (rs *ShareResource) find(c *gin.Context) {
	share, err := rs.dShare.FindByAlias(c.Param("alias"))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ginutil.JSONBadRequest(c, fmt.Errorf("share not found"))
		return
	} else if time.Now().After(share.ExpireAt) {
		ginutil.JSONForbidden(c, fmt.Errorf("share expired"))
		return
	}

	share.Secret = ""
	ginutil.JSONData(c, share)
}

func (rs *ShareResource) findAll(c *gin.Context) {
	p := new(bind.QueryPage)
	if err := c.BindQuery(p); err != nil {
		ginutil.JSONBadRequest(c, err)
		return
	}

	var total int64
	list := make([]model.Share, 0)
	sn := gormutil.DB().Where("uid=?", userIdGet(c))
	sn.Model(model.Share{}).Count(&total)
	sn = sn.Order("id desc")
	if err := sn.Limit(p.Limit).Offset(p.Offset).Find(&list).Error; err != nil {
		ginutil.JSONBadRequest(c, err)
		return
	}

	ginutil.JSONList(c, list, total)
}

func (rs *ShareResource) create(c *gin.Context) {
	p := new(bind.BodyShare)
	if err := c.ShouldBindJSON(p); err != nil {
		ginutil.JSONBadRequest(c, err)
		return
	}

	mMatter, err := rs.dMatter.Find(p.Matter)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ginutil.JSONBadRequest(c, fmt.Errorf("matter not found"))
		return
	}

	m := &model.Share{
		Alias:    strutil.RandomText(12),
		Uid:      userIdGet(c),
		Name:     mMatter.Name,
		Matter:   mMatter.Alias,
		Type:     mMatter.Type,
		ExpireAt: time.Now().Add(time.Second * time.Duration(p.ExpireSec)),
	}
	if p.Private {
		m.Secret = strutil.RandomText(5)
	}
	if err := gormutil.DB().Create(m).Error; err != nil {
		ginutil.JSONServerError(c, err)
		return
	}

	ginutil.JSONData(c, m)
}

func (rs *ShareResource) update(c *gin.Context) {
	p := new(bind.BodyShare)
	if err := c.ShouldBindJSON(p); err != nil {
		ginutil.JSONBadRequest(c, err)
		return
	}

	share, err := rs.dShare.Find(p.Id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ginutil.JSONBadRequest(c, fmt.Errorf("share not found"))
		return
	}

	if p.Private && share.Secret == "" {
		share.Secret = strutil.RandomText(5)
	}

	if err := gormutil.DB().Save(share).Error; err != nil {
		ginutil.JSONServerError(c, err)
		return
	}

	ginutil.JSON(c)
}

func (rs *ShareResource) delete(c *gin.Context) {
	share, err := rs.dShare.FindByAlias(c.Param("alias"))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ginutil.JSONBadRequest(c, fmt.Errorf("share not exist"))
		return
	}

	if err := gormutil.DB().Delete(share, "id=?", share.Id).Error; err != nil {
		ginutil.JSONServerError(c, err)
		return
	}

	ginutil.JSON(c)
}

func (rs *ShareResource) draw(c *gin.Context) {
	p := new(bind.BodyShareDraw)
	if err := c.ShouldBindJSON(p); err != nil {
		ginutil.JSONBadRequest(c, err)
		return
	}

	share, err := rs.dShare.FindByAlias(c.Param("alias"))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ginutil.JSONBadRequest(c, fmt.Errorf("share not exist"))
		return
	} else if share.Secret != p.Secret {
		ginutil.JSONForbidden(c, fmt.Errorf("invalid secret"))
		return
	}

	claims := &jwt.StandardClaims{
		ExpiresAt: share.ExpireAt.Unix(),
		IssuedAt:  time.Now().Unix(),
		NotBefore: time.Now().Unix(),
		Subject:   share.Alias,
	}
	token, err := rs.JWTUtil.Issue(claims)
	if err != nil {
		ginutil.JSONServerError(c, err)
		return
	}

	ginutil.Cookie(c, ShareCookieTokenKey, token, int(share.ExpireAt.Sub(time.Now()).Seconds()))
	ginutil.JSON(c)
}

func (rs *ShareResource) findMatter(c *gin.Context) {
	share, err := rs.dShare.FindByAlias(c.Param("alias"))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ginutil.JSONBadRequest(c, fmt.Errorf("share not exist"))
		return
	}

	if err := rs.shareTokenVerify(c, share); err != nil {
		ginutil.JSONBadRequest(c, err)
		return
	}

	mMatter, err := rs.dMatter.Find(share.Matter)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ginutil.JSONBadRequest(c, fmt.Errorf("matter not found"))
		return
	}

	ginutil.JSONData(c, mMatter)
}

func (rs *ShareResource) findMatters(c *gin.Context) {
	p := new(bind.QueryShareMatters)
	if err := c.ShouldBind(p); err != nil {
		ginutil.JSONBadRequest(c, err)
		return
	}

	share, err := rs.dShare.FindByAlias(c.Param("alias"))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ginutil.JSONBadRequest(c, fmt.Errorf("share not exist"))
		return
	}

	if err := rs.shareTokenVerify(c, share); err != nil {
		ginutil.JSONBadRequest(c, err)
		return
	}

	mMatter, err := rs.dMatter.Find(share.Matter)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ginutil.JSONBadRequest(c, fmt.Errorf("matter not found"))
		return
	}

	dir := fmt.Sprintf("%s/%s", mMatter.Name, p.Dir) // 设置父级目录
	list, total, err := rs.dMatter.FindAll(mMatter.Uid, p.Offset, p.Limit, matter.WithDir(dir))
	if err != nil {
		ginutil.JSONServerError(c, err)
		return
	}

	ginutil.JSONList(c, list, total)
}

func (rs *ShareResource) shareTokenVerify(c *gin.Context, share *model.Share) error {
	if !share.Protected {
		return nil
	}

	tokenStr, err := c.Cookie(ShareCookieTokenKey)
	if err != nil {
		return err
	}

	if token, err := rs.JWTUtil.Parse(tokenStr, &jwt.StandardClaims{}); err != nil {
		return err
	} else if token.Claims.(*jwt.StandardClaims).Subject != share.Alias {
		return fmt.Errorf("unmatched token")
	}

	return nil
}
