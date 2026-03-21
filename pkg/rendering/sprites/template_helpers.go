package sprites

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// entityRenderContext holds context for multi-phase entity rendering.
type entityRenderContext struct {
	img        *ebiten.Image
	config     Config
	entityType string
	direction  Direction
	genre      string
	useAerial  bool
	template   *AnatomicalTemplate
	traits     *AvatarTraits
}

// renderCreatureDetails renders creature-specific details for nonhumanoid entities.
func (ctx *entityRenderContext) renderCreatureDetails() {
	if !ctx.useAerial || IsHumanoidEntity(ctx.entityType) {
		return
	}

	detailBuf := image.NewRGBA(image.Rect(0, 0, ctx.config.Width, ctx.config.Height))
	RenderCreatureDetails(detailBuf, CreatureDetailParams{
		Width:     ctx.config.Width,
		Height:    ctx.config.Height,
		Form:      ctx.entityType,
		Direction: string(ctx.direction),
		Seed:      ctx.config.Seed,
		SizeClass: extractSizeClass(ctx.config),
		Genre:     ctx.genre,
	})
	detailImg := ebiten.NewImageFromImage(detailBuf)
	ctx.img.DrawImage(detailImg, nil)

	ctx.applyCreatureMarkings()
}

// applyCreatureMarkings applies seed-based creature markings (spots, stripes, etc).
func (ctx *entityRenderContext) applyCreatureMarkings() {
	creatureForm := EntityTypeToCreatureForm(ctx.entityType)
	markings := GenerateCreatureMarkings(ctx.config.Seed, creatureForm)
	if markings.Type == MarkingNone {
		return
	}

	markingBuf := image.NewRGBA(image.Rect(0, 0, ctx.config.Width, ctx.config.Height))
	if safeReadPixels(ctx.img, markingBuf.Pix) {
		RenderCreatureMarkings(markingBuf, CreatureMarkingParams{
			Width:     ctx.config.Width,
			Height:    ctx.config.Height,
			Form:      creatureForm,
			Direction: string(ctx.direction),
			Seed:      ctx.config.Seed,
			Markings:  markings,
		})
		markingImg := ebiten.NewImageFromImage(markingBuf)
		ctx.img.Clear()
		ctx.img.DrawImage(markingImg, nil)
	}
}

// renderGarmentDetails renders garment structure lines for humanoid entities.
func (ctx *entityRenderContext) renderGarmentDetails() {
	if !ctx.useAerial || !IsHumanoidEntity(ctx.entityType) {
		return
	}

	readBuf := image.NewRGBA(image.Rect(0, 0, ctx.config.Width, ctx.config.Height))
	if safeReadPixels(ctx.img, readBuf.Pix) {
		RenderGarmentDetail(readBuf, GarmentDetailParams{
			Width:     ctx.config.Width,
			Height:    ctx.config.Height,
			Seed:      ctx.config.Seed,
			Genre:     ctx.genre,
			Direction: string(ctx.direction),
		})
		garmentImg := ebiten.NewImageFromImage(readBuf)
		ctx.img.Clear()
		ctx.img.DrawImage(garmentImg, nil)
	}
}

// renderRoleDetails renders role-specific details for humanoid entities.
func (ctx *entityRenderContext) renderRoleDetails() {
	if !ctx.useAerial || !IsHumanoidEntity(ctx.entityType) {
		return
	}

	role := MapEntityTypeToRole(ctx.entityType)
	if role == "" {
		return
	}

	roleBuf := image.NewRGBA(image.Rect(0, 0, ctx.config.Width, ctx.config.Height))
	RenderRoleDetails(roleBuf, RoleDetailParams{
		Width:     ctx.config.Width,
		Height:    ctx.config.Height,
		Role:      role,
		Direction: string(ctx.direction),
		Seed:      ctx.config.Seed,
		Genre:     ctx.genre,
	})
	roleImg := ebiten.NewImageFromImage(roleBuf)
	ctx.img.DrawImage(roleImg, nil)
}

// renderBackAccessory renders back accessory overlay for humanoid entities.
func (ctx *entityRenderContext) renderBackAccessory() {
	if !ctx.useAerial || !IsHumanoidEntity(ctx.entityType) {
		return
	}

	torsoSpec, hasTorso := ctx.template.BodyPartLayout[PartTorso]
	if !hasTorso {
		return
	}

	baType := resolveBackAccessoryType(ctx.config)
	if baType == BackAccessoryNone {
		return
	}

	baParams := ComputeBackAccessoryParams(
		ctx.config.Width, ctx.config.Height, torsoSpec,
		baType, ctx.direction, ctx.config.Seed, ctx.genre,
	)
	baBuf := image.NewRGBA(image.Rect(0, 0, ctx.config.Width, ctx.config.Height))
	RenderBackAccessoryOverlay(baBuf, baParams)
	baImg := ebiten.NewImageFromImage(baBuf)
	ctx.img.DrawImage(baImg, nil)
}

// renderHeadgear renders headgear overlay for humanoid entities without equipment helmet.
func (ctx *entityRenderContext) renderHeadgear() {
	if !ctx.useAerial || !IsHumanoidEntity(ctx.entityType) || hasEquipmentHelmet(ctx.config) {
		return
	}

	headSpec, hasHead := ctx.template.BodyPartLayout[PartHead]
	if !hasHead {
		return
	}

	hgType := resolveHeadgearType(ctx.config)
	if hgType == HeadgearNone {
		return
	}

	hgParams := ComputeHeadgearParams(
		ctx.config.Width, ctx.config.Height, headSpec,
		hgType, ctx.direction, ctx.config.Seed, ctx.genre,
	)
	hgBuf := image.NewRGBA(image.Rect(0, 0, ctx.config.Width, ctx.config.Height))
	RenderHeadgearOverlay(hgBuf, hgParams)
	hgImg := ebiten.NewImageFromImage(hgBuf)
	ctx.img.DrawImage(hgImg, nil)
}

// applySurfaceTextures applies surface textures to nonhumanoid creatures.
func (ctx *entityRenderContext) applySurfaceTextures() {
	if !ctx.useAerial || IsHumanoidEntity(ctx.entityType) {
		return
	}

	form := EntityTypeToCreatureForm(ctx.entityType)
	texSet := GenerateSurfaceTextureSet(ctx.config.Seed, form, ctx.genre)
	if texSet.TorsoTexture.Type == TexNone {
		return
	}

	texBuf := image.NewRGBA(image.Rect(0, 0, ctx.config.Width, ctx.config.Height))
	if safeReadPixels(ctx.img, texBuf.Pix) {
		ApplySurfaceTexture(texBuf, texBuf.Bounds(), texSet.TorsoTexture, ctx.config.Seed)
		texImg := ebiten.NewImageFromImage(texBuf)
		ctx.img.Clear()
		ctx.img.DrawImage(texImg, nil)
	}
}

// applyDepthEnhancement applies volumetric depth enhancement.
func (ctx *entityRenderContext) applyDepthEnhancement() {
	if !ctx.useAerial {
		return
	}

	depthBuf := image.NewRGBA(image.Rect(0, 0, ctx.config.Width, ctx.config.Height))
	if safeReadPixels(ctx.img, depthBuf.Pix) {
		depthCfg := DefaultDepthEnhanceConfig(ctx.config.Seed)
		if !IsHumanoidEntity(ctx.entityType) {
			form := EntityTypeToCreatureForm(ctx.entityType)
			ApplyDepthEnhancementForCreature(depthBuf, form, depthCfg)
		} else {
			ApplyDepthEnhancement(depthBuf, depthCfg)
		}
		depthImg := ebiten.NewImageFromImage(depthBuf)
		ctx.img.Clear()
		ctx.img.DrawImage(depthImg, nil)
	}
}

// applyColorTemperature applies genre-aware color temperature grading.
func (ctx *entityRenderContext) applyColorTemperature() {
	if !ctx.useAerial {
		return
	}

	ctBuf := image.NewRGBA(image.Rect(0, 0, ctx.config.Width, ctx.config.Height))
	if safeReadPixels(ctx.img, ctBuf.Pix) {
		ctCfg := GenreColorTemperatureConfig(ctx.genre, ctx.config.Seed)
		ApplyColorTemperature(ctBuf, ctCfg)
		ctImg := ebiten.NewImageFromImage(ctBuf)
		ctx.img.Clear()
		ctx.img.DrawImage(ctImg, nil)
	}
}

// finalizeSprite applies sprite finalization: adaptive outline, rim lighting, edge shadow.
func (ctx *entityRenderContext) finalizeSprite() {
	if !ctx.useAerial {
		return
	}

	rgbaBuf := image.NewRGBA(image.Rect(0, 0, ctx.config.Width, ctx.config.Height))
	if safeReadPixels(ctx.img, rgbaBuf.Pix) {
		finalized := FinalizeEntitySprite(rgbaBuf, DefaultFinalizerConfig(ctx.config.Seed))
		finalImg := ebiten.NewImageFromImage(finalized)
		ctx.img.Clear()
		ctx.img.DrawImage(finalImg, nil)
	}
}
