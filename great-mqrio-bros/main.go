package main

import (
	"encoding/binary"
	"fmt"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth     = 800
	screenHeight    = 600
	stageWidth      = 2400
	playerWidth     = 32
	playerHeight    = 48
	gravity         = 0.5
	jumpPower       = -12
	moveSpeed       = 4
	audioSampleRate = 44100
)

// Player はプレイヤーキャラクターの構造体
type Player struct {
	x, y          float64 // 位置
	vx, vy        float64 // 速度（velocity）
	isGrounded    bool    // 地面に接しているか
	isFacingRight bool    // 右向きか
	animFrame     int     // アニメーションフレーム番号
	animCounter   int     // フレームカウンター
	state         string  // "idle", "walking", "jumping"
}

// Platform は足場の構造体
type Platform struct {
	x, y, width, height float64
	color               color.RGBA
}

// Enemy は敵キャラクターの構造体
type Enemy struct {
	x, y, width, height float64
	vx                  float64 // 左右移動速度
	leftBound           float64 // 移動範囲の左端
	rightBound          float64 // 移動範囲の右端
	isAlive             bool
	initialX            float64 // リセット用
	initialY            float64
	initialVx           float64
}

// Coin はコインの構造体
type Coin struct {
	x, y      float64
	radius    float64
	collected bool
}

// Goal はゴール（旗）の構造体
type Goal struct {
	x          float64
	y          float64
	poleHeight float64
	flagHeight float64 // 旗の現在位置（降りるアニメーション用）
	isReached  bool
}

// Game はゲームの状態を管理する構造体
type Game struct {
	player             Player
	platforms          []Platform
	enemies            []Enemy
	coins              []Coin
	goal               Goal
	gameState          string // "playing", "cleared"
	clearTime          int    // クリア後の経過フレーム数
	elapsedFrames      int    // プレイ開始からの経過フレーム数
	clearElapsedFrames int    // ゴール到達時点の経過フレーム（クリアタイム表示用）
	cameraX            float64
	score              int
	audioContext       *audio.Context
	jumpSound          *audio.Player
	coinSound          *audio.Player
	enemySound         *audio.Player
	goalSound          *audio.Player
}

// generateBeep は指定周波数・長さのサイン波を16bit LE ステレオPCMで返す
func generateBeep(sampleRate, durationMs, frequencyHz int) []byte {
	numSamples := sampleRate * durationMs / 1000
	// 2チャンネル（ステレオ）、各サンプル2バイト → 1サンプルあたり4バイト
	buf := make([]byte, numSamples*4)
	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		sample := math.Sin(2 * math.Pi * float64(frequencyHz) * t)
		sampleInt := int16(sample * 32767 * 0.3)
		u := uint16(sampleInt)
		// L, R 同じ値でステレオ
		binary.LittleEndian.PutUint16(buf[i*4:], u)
		binary.LittleEndian.PutUint16(buf[i*4+2:], u)
	}
	return buf
}

// NewGame は新しいゲームを作成
func NewGame() *Game {
	audioContext := audio.NewContext(audioSampleRate)
	jumpPCM := generateBeep(audioSampleRate, 100, 440)
	coinPCM := generateBeep(audioSampleRate, 150, 880)
	enemyPCM := generateBeep(audioSampleRate, 200, 220)

	goalPCM := generateBeep(audioSampleRate, 500, 660)

	jumpPlayer := audioContext.NewPlayerFromBytes(jumpPCM)
	coinPlayer := audioContext.NewPlayerFromBytes(coinPCM)
	enemyPlayer := audioContext.NewPlayerFromBytes(enemyPCM)
	goalPlayer := audioContext.NewPlayerFromBytes(goalPCM)

	return &Game{
		audioContext:       audioContext,
		jumpSound:          jumpPlayer,
		coinSound:          coinPlayer,
		enemySound:         enemyPlayer,
		goalSound:          goalPlayer,
		gameState:          "playing",
		elapsedFrames:      0,
		clearElapsedFrames: 0,
		goal: Goal{
			x:          stageWidth - 25, // 右端の浮き床(〜2350)より十分右に配置
			y:          450,
			poleHeight: 150,
			flagHeight: 0,
			isReached:  false,
		},
		player: Player{
			x:             100,
			y:             100,
			isFacingRight: true,
		},
		platforms: []Platform{
			// 地面（ステージ全体）
			{x: 0, y: 550, width: stageWidth, height: 50, color: color.RGBA{R: 100, G: 200, B: 100, A: 255}},
			// エリア1 (0-800)
			{x: 200, y: 450, width: 150, height: 20, color: color.RGBA{R: 139, G: 69, B: 19, A: 255}},
			{x: 400, y: 350, width: 150, height: 20, color: color.RGBA{R: 139, G: 69, B: 19, A: 255}},
			{x: 600, y: 450, width: 150, height: 20, color: color.RGBA{R: 139, G: 69, B: 19, A: 255}},
			{x: 350, y: 250, width: 100, height: 20, color: color.RGBA{R: 139, G: 69, B: 19, A: 255}},
			// エリア2 (800-1600)
			{x: 1000, y: 450, width: 150, height: 20, color: color.RGBA{R: 139, G: 69, B: 19, A: 255}},
			{x: 1200, y: 350, width: 150, height: 20, color: color.RGBA{R: 139, G: 69, B: 19, A: 255}},
			{x: 1400, y: 450, width: 150, height: 20, color: color.RGBA{R: 139, G: 69, B: 19, A: 255}},
			{x: 1150, y: 250, width: 100, height: 20, color: color.RGBA{R: 139, G: 69, B: 19, A: 255}},
			// エリア3 (1600-2400)
			{x: 1800, y: 450, width: 150, height: 20, color: color.RGBA{R: 139, G: 69, B: 19, A: 255}},
			{x: 2000, y: 350, width: 150, height: 20, color: color.RGBA{R: 139, G: 69, B: 19, A: 255}},
			{x: 2200, y: 450, width: 150, height: 20, color: color.RGBA{R: 139, G: 69, B: 19, A: 255}},
			{x: 1950, y: 250, width: 100, height: 20, color: color.RGBA{R: 139, G: 69, B: 19, A: 255}},
		},
		enemies: []Enemy{
			{x: 250, y: 426, width: 24, height: 24, vx: 2, leftBound: 200, rightBound: 326, isAlive: true, initialX: 250, initialY: 426, initialVx: 2},
			{x: 450, y: 326, width: 24, height: 24, vx: -2, leftBound: 400, rightBound: 526, isAlive: true, initialX: 450, initialY: 326, initialVx: -2},
			{x: 650, y: 426, width: 24, height: 24, vx: -2, leftBound: 600, rightBound: 726, isAlive: true, initialX: 650, initialY: 426, initialVx: -2},
			{x: 375, y: 226, width: 24, height: 24, vx: 1.5, leftBound: 350, rightBound: 426, isAlive: true, initialX: 375, initialY: 226, initialVx: 1.5},
			{x: 1050, y: 426, width: 24, height: 24, vx: -2, leftBound: 1000, rightBound: 1126, isAlive: true, initialX: 1050, initialY: 426, initialVx: -2},
			{x: 1250, y: 326, width: 24, height: 24, vx: 2, leftBound: 1200, rightBound: 1326, isAlive: true, initialX: 1250, initialY: 326, initialVx: 2},
			{x: 1450, y: 426, width: 24, height: 24, vx: -2, leftBound: 1400, rightBound: 1526, isAlive: true, initialX: 1450, initialY: 426, initialVx: -2},
			{x: 1175, y: 226, width: 24, height: 24, vx: 1.5, leftBound: 1150, rightBound: 1226, isAlive: true, initialX: 1175, initialY: 226, initialVx: 1.5},
			{x: 1850, y: 426, width: 24, height: 24, vx: 2, leftBound: 1800, rightBound: 1926, isAlive: true, initialX: 1850, initialY: 426, initialVx: 2},
			{x: 2050, y: 326, width: 24, height: 24, vx: -2, leftBound: 2000, rightBound: 2126, isAlive: true, initialX: 2050, initialY: 326, initialVx: -2},
			{x: 2250, y: 426, width: 24, height: 24, vx: -2, leftBound: 2200, rightBound: 2326, isAlive: true, initialX: 2250, initialY: 426, initialVx: -2},
			{x: 1975, y: 226, width: 24, height: 24, vx: 1.5, leftBound: 1950, rightBound: 2026, isAlive: true, initialX: 1975, initialY: 226, initialVx: 1.5},
		},
		coins: []Coin{
			{x: 150, y: 500, radius: 12, collected: false},
			{x: 280, y: 410, radius: 12, collected: false},
			{x: 350, y: 410, radius: 12, collected: false},
			{x: 480, y: 310, radius: 12, collected: false},
			{x: 520, y: 310, radius: 12, collected: false},
			{x: 680, y: 410, radius: 12, collected: false},
			{x: 400, y: 210, radius: 12, collected: false},
			{x: 250, y: 350, radius: 12, collected: false},
			{x: 550, y: 350, radius: 12, collected: false},
			{x: 400, y: 450, radius: 12, collected: false},
			{x: 100, y: 500, radius: 12, collected: false},
			{x: 950, y: 410, radius: 12, collected: false},
			{x: 1100, y: 310, radius: 12, collected: false},
			{x: 1300, y: 410, radius: 12, collected: false},
			{x: 1180, y: 210, radius: 12, collected: false},
			{x: 1750, y: 500, radius: 12, collected: false},
			{x: 1900, y: 410, radius: 12, collected: false},
			{x: 2100, y: 310, radius: 12, collected: false},
			{x: 2300, y: 500, radius: 12, collected: false},
		},
	}
}

// Update はゲームロジックを更新（毎フレーム呼ばれる）
func (g *Game) Update() error {
	if g.gameState == "cleared" {
		g.clearTime++
		if g.goal.flagHeight < g.goal.poleHeight-20 {
			g.goal.flagHeight += 2
		}
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			g.resetToStart()
		}
		return nil
	}

	// 左右移動の入力処理
	g.player.vx = 0
	if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		g.player.vx = -moveSpeed
		g.player.isFacingRight = false
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		g.player.vx = moveSpeed
		g.player.isFacingRight = true
	}

	// ジャンプの入力処理（地面にいる時のみ）
	if (ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyW)) && g.player.isGrounded {
		g.player.vy = jumpPower
		g.player.isGrounded = false
		if g.jumpSound != nil {
			_ = g.jumpSound.Rewind()
			g.jumpSound.Play()
		}
	}

	// 重力を適用
	g.player.vy += gravity

	// 速度の上限を設定（落下速度制限）
	if g.player.vy > 15 {
		g.player.vy = 15
	}

	// プレイヤーの状態とアニメーションを更新
	if !g.player.isGrounded {
		g.player.state = "jumping"
	} else if g.player.vx != 0 {
		g.player.state = "walking"
		g.player.animCounter++
		if g.player.animCounter >= 8 {
			g.player.animCounter = 0
			g.player.animFrame = 1 - g.player.animFrame
		}
	} else {
		g.player.state = "idle"
		g.player.animFrame = 0
		g.player.animCounter = 0
	}

	// プレイヤーの位置を更新
	g.player.x += g.player.vx
	g.player.y += g.player.vy

	// 衝突判定と位置補正
	g.checkCollisions()

	// コイン取得判定
	playerCenterX := g.player.x + playerWidth/2
	playerCenterY := g.player.y + playerHeight/2
	for i := range g.coins {
		if g.coins[i].collected {
			continue
		}
		dx := playerCenterX - g.coins[i].x
		dy := playerCenterY - g.coins[i].y
		distance := math.Sqrt(dx*dx + dy*dy)
		if distance < g.coins[i].radius+playerWidth/2 {
			g.coins[i].collected = true
			g.score += 10
			if g.coinSound != nil {
				_ = g.coinSound.Rewind()
				g.coinSound.Play()
			}
		}
	}

	// 敵の更新（左右移動、足場の端で折り返し）
	for i := range g.enemies {
		if !g.enemies[i].isAlive {
			continue
		}
		g.enemies[i].x += g.enemies[i].vx
		if g.enemies[i].x <= g.enemies[i].leftBound {
			g.enemies[i].x = g.enemies[i].leftBound
			g.enemies[i].vx = -g.enemies[i].vx
		}
		if g.enemies[i].x >= g.enemies[i].rightBound {
			g.enemies[i].x = g.enemies[i].rightBound
			g.enemies[i].vx = -g.enemies[i].vx
		}
	}

	// プレイヤーと敵の衝突判定
	for i := range g.enemies {
		if !g.enemies[i].isAlive {
			continue
		}
		playerLeft := g.player.x
		playerRight := g.player.x + playerWidth
		playerTop := g.player.y
		playerBottom := g.player.y + playerHeight
		enemyLeft := g.enemies[i].x
		enemyRight := g.enemies[i].x + g.enemies[i].width
		enemyTop := g.enemies[i].y
		enemyBottom := g.enemies[i].y + g.enemies[i].height

		if playerRight <= enemyLeft || playerLeft >= enemyRight ||
			playerBottom <= enemyTop || playerTop >= enemyBottom {
			continue
		}

		// 上から踏んだ場合: 敵を倒す、スコア+100、プレイヤーが小さくジャンプ
		if playerBottom < enemyTop+g.enemies[i].height/2 && g.player.vy > 0 {
			g.enemies[i].isAlive = false
			g.score += 100
			g.player.vy = -8 // 小さくジャンプ
			if g.enemySound != nil {
				_ = g.enemySound.Rewind()
				g.enemySound.Play()
			}
			continue
		}

		// 横から当たった場合: ゲームリセット
		g.resetGame()
		break
	}

	// カメラ追従: プレイヤーが画面中央より右にいたらカメラを追従
	targetX := g.player.x - screenWidth/2
	if targetX < 0 {
		targetX = 0
	}
	if targetX > stageWidth-screenWidth {
		targetX = stageWidth - screenWidth
	}
	g.cameraX = targetX

	// ステージの左右端でプレイヤーを止める
	if g.player.x < 0 {
		g.player.x = 0
	}
	if g.player.x > stageWidth-playerWidth {
		g.player.x = stageWidth - playerWidth
	}

	// ゴール判定
	if !g.goal.isReached &&
		g.player.x+playerWidth >= g.goal.x &&
		g.player.x <= g.goal.x+30 {
		g.goal.isReached = true
		g.clearElapsedFrames = g.elapsedFrames
		g.gameState = "cleared"
		g.clearTime = 0
		remainingCoins := 0
		for _, c := range g.coins {
			if !c.collected {
				remainingCoins++
			}
		}
		g.score += remainingCoins * 50
		if g.goalSound != nil {
			_ = g.goalSound.Rewind()
			g.goalSound.Play()
		}
	}

	// 画面下に落ちたらリセット
	if g.player.y > screenHeight {
		g.resetGame()
	}

	g.elapsedFrames++
	return nil
}

// resetGame はプレイヤーをスタート地点に戻し、敵とコインを復活させる（落下時など）
func (g *Game) resetGame() {
	g.resetToStart()
}

// resetToStart はリスタート用。音声コンテキスト・Player はそのまま使い、ゲーム状態だけ初期化する。
func (g *Game) resetToStart() {
	g.gameState = "playing"
	g.clearTime = 0
	g.elapsedFrames = 0
	g.clearElapsedFrames = 0
	g.player.x = 100
	g.player.y = 100
	g.player.vx = 0
	g.player.vy = 0
	g.player.isGrounded = false
	g.player.state = "idle"
	g.player.animFrame = 0
	g.player.animCounter = 0
	g.cameraX = 0
	g.score = 0

	g.goal.isReached = false
	g.goal.flagHeight = 0

	for i := range g.enemies {
		g.enemies[i].isAlive = true
		g.enemies[i].x = g.enemies[i].initialX
		g.enemies[i].y = g.enemies[i].initialY
		g.enemies[i].vx = g.enemies[i].initialVx
	}
	for i := range g.coins {
		g.coins[i].collected = false
	}
}

// checkCollisions は衝突判定を行う
func (g *Game) checkCollisions() {
	g.player.isGrounded = false

	playerLeft := g.player.x
	playerRight := g.player.x + playerWidth
	playerTop := g.player.y
	playerBottom := g.player.y + playerHeight

	for _, platform := range g.platforms {
		platLeft := platform.x
		platRight := platform.x + platform.width
		platTop := platform.y
		platBottom := platform.y + platform.height

		// 重なっているか判定
		if playerRight > platLeft && playerLeft < platRight &&
			playerBottom > platTop && playerTop < platBottom {

			// 下から衝突（頭をぶつける）
			if g.player.vy < 0 && playerTop < platBottom && playerBottom > platBottom {
				g.player.y = platBottom
				g.player.vy = 0
			}

			// 上から衝突（着地）
			if g.player.vy > 0 && playerBottom > platTop && playerTop < platTop {
				g.player.y = platTop - playerHeight
				g.player.vy = 0
				g.player.isGrounded = true
			}

			// 左から衝突
			if g.player.vx > 0 && playerRight > platLeft && playerLeft < platLeft {
				g.player.x = platLeft - playerWidth
			}

			// 右から衝突
			if g.player.vx < 0 && playerLeft < platRight && playerRight > platRight {
				g.player.x = platRight
			}
		}
	}
}

// Draw は画面に描画（毎フレーム呼ばれる）
func (g *Game) Draw(screen *ebiten.Image) {
	// 背景（空）
	screen.Fill(color.RGBA{R: 135, G: 206, B: 235, A: 255})

	// カメラオフセットを適用して描画
	cam := float32(g.cameraX)

	// 足場を描画
	for _, platform := range g.platforms {
		vector.DrawFilledRect(
			screen,
			float32(platform.x)-cam,
			float32(platform.y),
			float32(platform.width),
			float32(platform.height),
			platform.color,
			false,
		)
	}

	// コインを描画（黄色い円）
	coinColor := color.RGBA{R: 255, G: 215, B: 0, A: 255}
	for _, coin := range g.coins {
		if coin.collected {
			continue
		}
		vector.DrawFilledCircle(
			screen,
			float32(coin.x)-cam,
			float32(coin.y),
			float32(coin.radius),
			coinColor,
			false,
		)
	}

	// ゴールを描画（ポール＋旗）
	poleColor := color.RGBA{R: 100, G: 100, B: 100, A: 255}
	vector.DrawFilledRect(
		screen,
		float32(g.goal.x)-cam,
		float32(g.goal.y),
		8,
		float32(g.goal.poleHeight),
		poleColor,
		false,
	)
	flagColor := color.RGBA{R: 255, G: 50, B: 50, A: 255}
	flagY := g.goal.y + g.goal.flagHeight
	vector.DrawFilledRect(
		screen,
		float32(g.goal.x+8)-cam,
		float32(flagY),
		24,
		16,
		flagColor,
		false,
	)

	// 敵を描画（茶色い四角形）
	enemyColor := color.RGBA{R: 139, G: 90, B: 43, A: 255}
	for _, enemy := range g.enemies {
		if !enemy.isAlive {
			continue
		}
		vector.DrawFilledRect(
			screen,
			float32(enemy.x)-cam,
			float32(enemy.y),
			float32(enemy.width),
			float32(enemy.height),
			enemyColor,
			false,
		)
	}

	// プレイヤーを描画（体）- 状態に応じた高さ
	bodyHeight := playerHeight
	bodyYOffset := 0.0
	switch g.player.state {
	case "walking":
		if g.player.animFrame == 1 {
			bodyHeight -= 2
			bodyYOffset = 2 // 片足を上げた表現
		}
	case "jumping":
		bodyHeight -= 4
		bodyYOffset = 2
	}
	playerColor := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	vector.DrawFilledRect(
		screen,
		float32(g.player.x)-cam,
		float32(g.player.y+bodyYOffset),
		playerWidth,
		float32(bodyHeight),
		playerColor,
		false,
	)

	// プレイヤーの顔（白い部分）
	vector.DrawFilledRect(
		screen,
		float32(g.player.x+8)-cam,
		float32(g.player.y+8+bodyYOffset),
		16,
		16,
		color.RGBA{R: 255, G: 220, B: 177, A: 255},
		false,
	)

	// 向きを示す矢印
	faceY := g.player.y + 14 + bodyYOffset
	if g.player.isFacingRight {
		vector.DrawFilledRect(
			screen,
			float32(g.player.x+20)-cam,
			float32(faceY),
			8,
			6,
			color.RGBA{R: 0, G: 0, B: 0, A: 255},
			false,
		)
	} else {
		vector.DrawFilledRect(
			screen,
			float32(g.player.x+4)-cam,
			float32(faceY),
			8,
			6,
			color.RGBA{R: 0, G: 0, B: 0, A: 255},
			false,
		)
	}

	// Controls and status
	status := fmt.Sprintf(
		"Controls: ←→ or A/D = move, SPACE or ↑ or W = jump\n"+
			"Pos: (%.0f, %.0f) Vel: (%.1f, %.1f) Grounded: %v\n"+
			"Score: %d",
		g.player.x, g.player.y, g.player.vx, g.player.vy, g.player.isGrounded,
		g.score,
	)
	ebitenutil.DebugPrint(screen, status)

	// クリア画面
	if g.gameState == "cleared" {
		overlay := ebiten.NewImage(screenWidth, screenHeight)
		overlay.Fill(color.RGBA{R: 0, G: 0, B: 0, A: 180})
		op := &ebiten.DrawImageOptions{}
		screen.DrawImage(overlay, op)
		clearTimeSec := float64(g.clearElapsedFrames) / 60.0
		remainingCoins := 0
		for _, c := range g.coins {
			if !c.collected {
				remainingCoins++
			}
		}
		coinBonus := remainingCoins * 50
		msg := fmt.Sprintf(
			"STAGE CLEAR!\n\n"+
				"Time: %.2f sec\n"+
				"Score: %d\n"+
				"Coin bonus: %d\n\n"+
				"[ RESTART ] SPACE",
			clearTimeSec, g.score, coinBonus,
		)
		ebitenutil.DebugPrintAt(screen, msg, screenWidth/2-100, screenHeight/2-60)
	}
}

// Layout は画面サイズを返す
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	// ウィンドウの設定
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Mario-style Platformer - Ebitengine")

	// ゲームを開始
	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
