package SpaceCookies

import (
	"fmt"
	. "github.com/vova616/GarageEngine/Engine"
	. "github.com/vova616/GarageEngine/Engine/Components"
	_ "image/jpeg"
	_ "image/png"
	//"gl"  
	"strconv"
	"time"
	//"strings"
	//"math"
	c "github.com/vova616/chipmunk"
	. "github.com/vova616/chipmunk/vect"
	//"image"
	"math/rand"
)

type GameScene struct {
	*SceneData
	Layer1 *GameObject
	Layer2 *GameObject
	Layer3 *GameObject
	Layer4 *GameObject
}

var (
	GameSceneGeneral *GameScene
	cir              *Texture
	boxt             *Texture
	cookie           *GameObject
	defender         *GameObject

	Player     *GameObject
	PlayerShip *ShipController

	Explosion *GameObject
	PowerUpGO *GameObject

	Wall *GameObject

	atlas        = NewManagedAtlas(2048, 1024)
	atlasSpace   = NewManagedAtlas(1024, 1024)
	atlasPowerUp = NewManagedAtlas(256, 256)
	backgroung   *Texture

	queenDead = false
)

const (
	MissleTag = "Missle"
	CookieTag = "Cookie"
)
const SpaceShip_A = 333
const Missle_A = 334
const HP_A = 123
const HPGUI_A = 124
const Queen_A = 666
const Jet_A = 125

func CheckError(err error) bool {
	if err != nil {
		fmt.Println(err)
		return true
	}
	return false
}

func LoadTextures() {
	CheckError(atlas.LoadImage("./data/SpaceCookies/Ship1.png", SpaceShip_A))
	CheckError(atlas.LoadImage("./data/SpaceCookies/missile.png", Missle_A))
	CheckError(atlas.LoadGroupSheet("./data/SpaceCookies/Explosion.png", 128, 128, 6*8))

	CheckError(atlas.LoadImage("./data/SpaceCookies/HealthBar.png", HP_A))
	CheckError(atlas.LoadImage("./data/SpaceCookies/HealthBarGUI.png", HPGUI_A))
	CheckError(atlas.LoadImage("./data/SpaceCookies/Queen.png", Queen_A))
	CheckError(atlas.LoadImage("./data/SpaceCookies/Jet.png", Jet_A))

	atlas.BuildAtlas()
	atlas.BuildMipmaps()
	atlas.SetFiltering(MinFilter, MipMapLinearNearest)
	atlas.SetFiltering(MagFilter, Nearest)
	atlas.Texture.SetReadOnly()

	var e error
	boxt, e = LoadTexture("./data/SpaceCookies/wall.png")

	boxt.BuildMipmaps()
	boxt.SetFiltering(MinFilter, MipMapLinearNearest)
	boxt.SetFiltering(MagFilter, Nearest)

	backgroung, e = LoadTexture("./data/SpaceCookies/background.png")
	CheckError(e)
	cir, e = LoadTexture("./data/SpaceCookies/Cookie.png")
	CheckError(e)

	cir.BuildMipmaps()
	cir.SetFiltering(MinFilter, MipMapLinearNearest)
	cir.SetFiltering(MagFilter, Nearest)

	CheckError(atlasSpace.LoadGroup("./data/SpaceCookies/Space/"))
	atlasSpace.BuildAtlas()
	atlasSpace.BuildMipmaps()
	atlasSpace.SetFiltering(MinFilter, MipMapLinearNearest)
	atlasSpace.SetFiltering(MagFilter, Nearest)
	atlasSpace.Texture.SetReadOnly()

	CheckError(atlasPowerUp.LoadGroupSheet("./data/SpaceCookies/powerups.png", 61, 61, 3*4))
	atlasPowerUp.BuildAtlas()
	atlasPowerUp.SetFiltering(MinFilter, Linear)
	atlasPowerUp.SetFiltering(MagFilter, Linear)
}

func init() {
	Title = "Space Cookies"
}

func (s *GameScene) Load() {

	queenDead = false

	rand.Seed(time.Now().UnixNano())

	ArialFont, err := NewFont("./data/Fonts/arial.ttf", 48)
	if err != nil {
		panic(err)
	}
	ArialFont.Texture.SetReadOnly()

	ArialFont2, err := NewFont("./data/Fonts/arial.ttf", 24)
	if err != nil {
		panic(err)
	}
	ArialFont2.Texture.SetReadOnly()

	_ = ArialFont2
	_ = ArialFont

	GameSceneGeneral = s

	s.Camera = NewCamera()

	cam := NewGameObject("Camera")
	cam.AddComponent(s.Camera)

	cam.Transform().SetScale(NewVector3(1, 1, 1))

	gui := NewGameObject("GUI")

	Layer1 := NewGameObject("Layer1")
	Layer2 := NewGameObject("Layer2")
	Layer3 := NewGameObject("Layer3")
	Layer4 := NewGameObject("Layer3")

	s.Layer1 = Layer1
	s.Layer2 = Layer2
	s.Layer3 = Layer3
	s.Layer4 = Layer4

	mouse := NewGameObject("Mouse")
	mouse.AddComponent(NewMouse())
	mouse.AddComponent(NewMouseDebugger())
	mouse.Transform().SetParent2(cam)

	FPSDrawer := NewGameObject("FPS")
	FPSDrawer.Transform().SetParent2(cam)
	txt := FPSDrawer.AddComponent(NewUIText(ArialFont2, "")).(*UIText)
	fps := FPSDrawer.AddComponent(NewFPS()).(*FPS)
	fps.SetAction(func(fps float32) {
		txt.SetString("FPS: " + strconv.FormatFloat(float64(fps), 'f', 2, 32))
	})

	FPSDrawer.Transform().SetPosition(NewVector2(60, float32(Height)-20))
	FPSDrawer.Transform().SetScale(NewVector2(20, 20))

	//SPACCCEEEEE
	Space.Gravity.Y = 0
	Space.Iterations = 10

	Health := NewGameObject("HP")
	Health.Transform().SetParent2(cam)
	Health.Transform().SetPosition(NewVector2(150, 50))

	HealthGUI := NewGameObject("HPGUI")
	HealthGUI.AddComponent(NewSprite2(atlas.Texture, IndexUV(atlas, HPGUI_A)))
	HealthGUI.Transform().SetParent2(Health)
	HealthGUI.Transform().SetScale(NewVector2(50, 50))

	HealthBar := NewGameObject("HealthBar")
	HealthBar.Transform().SetParent2(Health)
	HealthBar.Transform().SetPosition(NewVector2(-82, 0))
	HealthBar.Transform().SetScale(NewVector2(100, 50))

	uvHP := IndexUV(atlas, HP_A)

	HealthBarGUI := NewGameObject("HealthBarGUI")
	HealthBarGUI.Transform().SetParent2(HealthBar)
	HealthBarGUI.AddComponent(NewSprite2(atlas.Texture, uvHP))
	HealthBarGUI.Transform().SetScale(NewVector2(0.52, 1))
	HealthBarGUI.Transform().SetPosition(NewVector2((uvHP.Ratio/2)*HealthBarGUI.Transform().Scale().X, 0))

	JetFire := NewGameObject("Jet")
	JetFire.AddComponent(NewSprite2(atlas.Texture, IndexUV(atlas, Jet_A)))

	ship := NewGameObject("Ship")
	Player = ship
	ship.AddComponent(NewSprite2(atlas.Texture, IndexUV(atlas, SpaceShip_A)))
	PlayerShip = ship.AddComponent(NewShipController()).(*ShipController)
	ship.Transform().SetParent2(Layer2)
	ship.Transform().SetPosition(NewVector2(400, 200))
	ship.Transform().SetScale(NewVector2(100, 100))
	ship.AddComponent(NewDestoyable(1000, 1))
	PlayerShip.HPBar = HealthBar
	PlayerShip.JetFire = JetFire

	uvs, ind := AnimatedGroupUVs(atlas, "Explosion")
	Explosion = NewGameObject("Explosion")
	Explosion.AddComponent(NewSprite3(atlas.Texture, uvs))
	Explosion.Sprite.BindAnimations(ind)
	Explosion.Sprite.AnimationSpeed = 20
	Explosion.Sprite.AnimationEndCallback = func(sprite *Sprite) {
		sprite.GameObject().Destroy()
	}
	Explosion.Transform().SetScale(NewVector2(30, 30))

	missle := NewGameObject("Missle")
	missle.AddComponent(NewSprite2(atlas.Texture, IndexUV(atlas, Missle_A)))
	missle.AddComponent(NewPhysics(false, 10, 10))
	missle.Transform().SetScale(NewVector2(20, 20))
	missle.AddComponent(NewDamageDealer(50))

	m := NewMissle(30000)
	missle.AddComponent(m)
	PlayerShip.Missle = m
	m.Explosion = Explosion
	ds := NewDestoyable(0, 1)
	ds.SetDestroyTime(1)
	missle.AddComponent(ds)

	cookie = NewGameObject("Cookie")
	cookie.AddComponent(NewSprite(cir))
	cookie.AddComponent(NewDestoyable(100, 2))
	cookie.AddComponent(NewDamageDealer(20))
	cookie.AddComponent(NewEnemeyAI(Player, Enemey_Cookie))
	cookie.Transform().SetScale(NewVector2(50, 50))
	cookie.Transform().SetPosition(NewVector2(400, 400))
	cookie.AddComponent(NewPhysics2(false, c.NewCircle(Vect{0, 0}, 25)))
	cookie.Tag = CookieTag

	defender = NewGameObject("Box")
	ds = NewDestoyable(30, 3)
	ds.SetDestroyTime(5)
	defender.AddComponent(ds)
	defender.AddComponent(NewSprite(boxt))
	defender.Tag = CookieTag
	defender.Transform().SetScale(NewVector2(50, 50))

	phx := defender.AddComponent(NewPhysics(false, 50, 50)).(*Physics)
	phx.Shape.SetFriction(0.5)
	//phx.Shape.Group = 2
	phx.Shape.SetElasticity(0.5)

	QueenCookie := NewGameObject("Cookie")
	QueenCookie.AddComponent(NewSprite2(atlas.Texture, IndexUV(atlas, Queen_A)))
	QueenCookie.AddComponent(NewDestoyable(5000, 2))
	QueenCookie.AddComponent(NewDamageDealer(200))
	QueenCookie.AddComponent(NewEnemeyAI(Player, Enemey_Boss))
	QueenCookie.Transform().SetParent2(Layer2)
	QueenCookie.Transform().SetScale(NewVector2(300, 300))
	QueenCookie.Transform().SetPosition(NewVector2(999999, 999999))
	QueenCookie.AddComponent(NewPhysics2(false, c.NewCircle(Vect{0, 0}, 25)))
	QueenCookie.Tag = CookieTag

	staticCookie := NewGameObject("Cookie")
	staticCookie.AddComponent(NewSprite(cir))
	staticCookie.Transform().SetScale(NewVector2(400, 400))
	staticCookie.Transform().SetPosition(NewVector2(400, 400))
	staticCookie.AddComponent(NewDestoyable(float32(Inf), 2))
	staticCookie.AddComponent(NewPhysics2(true, c.NewCircle(Vect{0, 0}, 200)))

	staticCookie.Physics.Shape.SetElasticity(0)
	staticCookie.Physics.Body.SetMass(999999999999)
	staticCookie.Physics.Body.SetMoment(staticCookie.Physics.Shape.Moment(999999999999))
	staticCookie.Tag = CookieTag

	uvs, ind = AnimatedGroupUVs(atlasSpace, "s")
	Background := NewGameObject("Background")
	Background.AddComponent(NewSprite3(atlasSpace.Texture, uvs))
	Background.Sprite.BindAnimations(ind)
	Background.Sprite.SetAnimation("s")
	Background.Sprite.AnimationSpeed = 0
	Background.Transform().SetScale(NewVector2(50, 50))
	Background.Transform().SetPosition(NewVector2(400, 400))

	uvs, ind = AnimatedGroupUVs(atlasPowerUp, "powerups")
	PowerUpGO = NewGameObject("Background")
	//PowerUpGO.Transform().SetParent2(Layer2)
	PowerUpGO.AddComponent(NewSprite3(atlasPowerUp.Texture, uvs))
	PowerUpGO.AddComponent(NewPhysics(false, 61, 61))
	PowerUpGO.Physics.Shape.IsSensor = true
	PowerUpGO.Sprite.BindAnimations(ind)
	PowerUpGO.Sprite.SetAnimation("powerups")
	PowerUpGO.Sprite.AnimationSpeed = 0
	index := (rand.Int() % 6) + 6
	PowerUpGO.Sprite.SetAnimationIndex(int(index))
	PowerUpGO.Transform().SetScale(NewVector2(61, 61))
	PowerUpGO.Transform().SetPosition(NewVector2(0, 0))

	background := NewGameObject("Background")
	background.AddComponent(NewSprite(backgroung))
	background.AddComponent(NewBackground(background.Sprite))
	background.Sprite.Render = false
	//background.Transform().SetScalef(float32(backgroung.Height()), float32(backgroung.Height()), 1)
	background.Transform().SetScalef(800, 800, 1)
	background.Transform().SetPositionf(0, 0, 0)

	for i := 0; i < 300; i++ {
		c := Background.Clone()
		c.Transform().SetParent2(Layer4)
		size := 20 + rand.Float32()*50
		p := Vector{(rand.Float32() * 5000) - 1000, (rand.Float32() * 5000) - 1000, 1}

		index := rand.Int() % 7

		Background.Sprite.SetAnimationIndex(int(index))

		c.Transform().SetRotationf(0, 0, rand.Float32()*360)

		c.Transform().SetPosition(p)
		c.Transform().SetScalef(size, size, 1)
	}

	for i := 0; i < 600; i++ {
		c := cookie.Clone()
		//c.Tag = CookieTag
		c.Transform().SetParent2(Layer2)
		size := 40 + rand.Float32()*100
		p := Vector{(rand.Float32() * 4000), (rand.Float32() * 4000), 1}

		if p.X < 1100 && p.Y < 800 {
			p.X += 1100
			p.Y += 800
		}

		c.Transform().SetPosition(p)
		c.Transform().SetScalef(size, size, 1)
	}

	Wall = NewGameObject("Wall")
	Wall.Transform().SetParent2(Layer2)

	for i := 0; i < (4000/400)+2; i++ {
		c := staticCookie.Clone()
		c.Transform().SetParent2(Wall)
		p := Vector{float32(i) * 400, -200, 1}
		c.Transform().SetPosition(p)
		c.Transform().SetScalef(400, 400, 1)
	}
	for i := 0; i < (4000/400)+2; i++ {
		c := staticCookie.Clone()
		c.Transform().SetParent2(Wall)
		p := Vector{float32(i) * 400, 4200, 1}
		c.Transform().SetPosition(p)
		c.Transform().SetScalef(400, 400, 1)
	}
	for i := 0; i < (4000/400)+2; i++ {
		c := staticCookie.Clone()
		c.Transform().SetParent2(Wall)
		p := Vector{-200, float32(i) * 400, 1}
		c.Transform().SetPosition(p)
		c.Transform().SetScalef(400, 400, 1)
	}
	for i := 0; i < (4000/400)+2; i++ {
		c := staticCookie.Clone()
		c.Transform().SetParent2(Wall)
		p := Vector{4200, float32(i) * 400, 1}
		c.Transform().SetPosition(p)
		c.Transform().SetScalef(400, 400, 1)
	}

	s.AddGameObject(cam)
	s.AddGameObject(gui)
	s.AddGameObject(Layer1)
	s.AddGameObject(Layer2)
	s.AddGameObject(Layer3)
	s.AddGameObject(Layer4)
	s.AddGameObject(background)
	//s.AddGameObject(shadowShader)

	fmt.Println("Scene loaded")
}

func (s *GameScene) SceneBase() *SceneData {
	return s.SceneData
}

func (s *GameScene) New() Scene {
	gs := new(GameScene)
	gs.SceneData = NewScene("GameScene")
	return gs
}