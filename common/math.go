package common

type Vector2F struct {
	X, Y float32
}

func NewVector2I(x, y int) Vector2F {
	return Vector2F{
		X: float32(x),
		Y: float32(y),
	}
}

func NewVector2F(x, y float32) Vector2F {
	return Vector2F{
		X: x,
		Y: y,
	}
}

func (v Vector2F) AddVector2F(other Vector2F) Vector2F {
	return Vector2F{
		X: v.X + other.X,
		Y: v.Y + other.Y,
	}
}

func (v Vector2F) AddScalar(num float32) Vector2F {
	return Vector2F{
		X: v.X + num,
		Y: v.Y + num,
	}
}

func (v Vector2F) MutliplyVector2F(other Vector2F) Vector2F {
	return Vector2F{
		X: v.X * other.X,
		Y: v.Y * other.Y,
	}
}

func (v Vector2F) MutliplyScalar(num float32) Vector2F {
	return Vector2F{
		X: v.X * num,
		Y: v.Y * num,
	}
}

func (v Vector2F) Int() (int, int) {
	return int(v.X), int(v.Y)
}
func (v Vector2F) Equals(other Vector2F) bool {
	return v.X == other.X && v.Y == other.Y
}

type Rect struct {
	X, Y, W, H float32
}

func (r *Rect) IsZero() bool {
	return r.X == 0 && r.Y == 0 && r.W == 0 && r.H == 0
}

func (r Rect) Equals(other Rect) bool {
	return r.X == other.X && r.Y == other.Y && r.H == other.H && r.W == other.W
}
