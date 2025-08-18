package memory

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/whynot00/go-telegram-fsm/media"
)

func mf(tpe, id string) media.File { return media.File{Type: tpe, FileID: id} }

// --- базовые микро-бенчи ---

func BenchmarkMemoryStorage_Set(b *testing.B) {
	store := NewMemoryStorage(30*time.Second, 30*time.Minute)
	ctx := context.Background()
	userID := int64(1)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Set(ctx, userID, "k", i)
	}
}

func BenchmarkMemoryStorage_Get_Hit(b *testing.B) {
	store := NewMemoryStorage(30*time.Second, 30*time.Minute)
	ctx := context.Background()
	userID := int64(1)
	store.Set(ctx, userID, "k", 42)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if v, ok := store.Get(ctx, userID, "k"); !ok || v.(int) == -1 {
			b.Fail()
		}
	}
}

func BenchmarkMemoryStorage_Get_Miss(b *testing.B) {
	store := NewMemoryStorage(30*time.Second, 30*time.Minute)
	ctx := context.Background()
	userID := int64(1)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, ok := store.Get(ctx, userID, "absent"); ok {
			b.Fail()
		}
	}
}

func BenchmarkMemoryStorage_SetMedia(b *testing.B) {
	store := NewMemoryStorage(30*time.Second, 30*time.Minute)
	ctx := context.Background()
	userID := int64(1)
	groupID := "g"

	file := mf("photo", "id")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.SetMedia(ctx, userID, groupID, file)
	}
}

func BenchmarkMemoryStorage_GetMedia_Hit(b *testing.B) {
	store := NewMemoryStorage(30*time.Second, 30*time.Minute)
	ctx := context.Background()
	userID := int64(1)
	groupID := "g"
	store.SetMedia(ctx, userID, groupID, mf("photo", "x"))

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, ok := store.GetMedia(ctx, userID, groupID); !ok {
			b.Fail()
		}
	}
}

func BenchmarkMemoryStorage_GetMedia_Miss(b *testing.B) {
	store := NewMemoryStorage(30*time.Second, 30*time.Minute)
	ctx := context.Background()
	userID := int64(1)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, ok := store.GetMedia(ctx, userID, "absent"); ok {
			b.Fail()
		}
	}
}

// --- конкуренция: запись в один и тот же mediaGroup ---

func BenchmarkMemoryStorage_SetMedia_SingleHotGroup_Parallel(b *testing.B) {
	store := NewMemoryStorage(30*time.Second, 30*time.Minute)
	ctx := context.Background()
	userID := int64(1)
	groupID := "hot"
	file := mf("photo", "id")

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			store.SetMedia(ctx, userID, groupID, file)
		}
	})
}

// параллельное чтение одного mediaGroupID
func BenchmarkMemoryStorage_GetMedia_SingleHotGroup_Parallel(b *testing.B) {
	store := NewMemoryStorage(30*time.Second, 30*time.Minute)
	ctx := context.Background()
	userID := int64(1)
	groupID := "hot"
	// прогреем
	for i := 0; i < 1000; i++ {
		store.SetMedia(ctx, userID, groupID, mf("photo", strconv.Itoa(i)))
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, ok := store.GetMedia(ctx, userID, groupID); !ok {
				b.Fail()
			}
		}
	})
}

// смешанная параллель: 50% SetMedia, 50% GetMedia
func BenchmarkMemoryStorage_Mixed_Parallel(b *testing.B) {
	store := NewMemoryStorage(30*time.Second, 30*time.Minute)
	ctx := context.Background()
	userID := int64(1)
	groupID := "hot"
	file := mf("video", "id")

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i&1 == 0 {
				store.SetMedia(ctx, userID, groupID, file)
			} else {
				store.GetMedia(ctx, userID, groupID)
			}
			i++
		}
	})
}

// --- масштабирование: U пользователей × G групп ---

func BenchmarkMemoryStorage_Scale_SetMedia(b *testing.B) {
	type cfg struct{ users, groups int }
	cases := []cfg{
		{1, 1}, {1, 16}, {10, 16}, {100, 16}, {100, 64},
	}
	for _, c := range cases {
		name := "U" + strconv.Itoa(c.users) + "_G" + strconv.Itoa(c.groups)
		b.Run(name, func(b *testing.B) {
			store := NewMemoryStorage(30*time.Second, 30*time.Minute)
			ctx := context.Background()
			files := make([]media.File, 64)
			for i := range files {
				files[i] = mf("photo", strconv.Itoa(i))
			}
			var uIDs []int64
			for u := 0; u < c.users; u++ {
				uIDs = append(uIDs, int64(u+1))
			}
			var groups []string
			for g := 0; g < c.groups; g++ {
				groups = append(groups, "g"+strconv.Itoa(g))
			}

			b.ReportAllocs()
			b.ResetTimer()
			idx := 0
			for i := 0; i < b.N; i++ {
				u := uIDs[i%len(uIDs)]
				g := groups[i%len(groups)]
				store.SetMedia(ctx, u, g, files[idx%len(files)])
				idx++
			}
		})
	}
}

func BenchmarkMemoryStorage_Scale_GetMedia(b *testing.B) {
	type cfg struct{ users, groups int }
	cases := []cfg{
		{1, 1}, {1, 16}, {10, 16}, {100, 16}, {100, 64},
	}
	for _, c := range cases {
		name := "U" + strconv.Itoa(c.users) + "_G" + strconv.Itoa(c.groups)
		b.Run(name, func(b *testing.B) {
			store := NewMemoryStorage(30*time.Second, 30*time.Minute)
			ctx := context.Background()
			// сетап: наполним
			for u := 0; u < c.users; u++ {
				for g := 0; g < c.groups; g++ {
					for k := 0; k < 8; k++ {
						store.SetMedia(ctx, int64(u+1), "g"+strconv.Itoa(g), mf("photo", strconv.Itoa(k)))
					}
				}
			}
			var uIDs []int64
			for u := 0; u < c.users; u++ {
				uIDs = append(uIDs, int64(u+1))
			}
			var groups []string
			for g := 0; g < c.groups; g++ {
				groups = append(groups, "g"+strconv.Itoa(g))
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				u := uIDs[i%len(uIDs)]
				g := groups[i%len(groups)]
				if _, ok := store.GetMedia(ctx, u, g); !ok {
					b.Fail()
				}
			}
		})
	}
}

// --- чистки ---

func BenchmarkMemoryStorage_CleanCache_Batched(b *testing.B) {
	store := NewMemoryStorage(30*time.Second, 30*time.Minute)
	ctx := context.Background()
	const groups, K = 32, 256
	const userID = int64(7)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if i%K == 0 {
			// Пакетное наполнение K раз — тоже внутри измерения.
			for j := 0; j < K; j++ {
				store.Set(ctx, userID, "k", 1)
				for g := 0; g < groups; g++ {
					store.SetMedia(ctx, userID, "g"+strconv.Itoa(g), mf("photo", "x"))
				}
			}
		}
		store.CleanCache(ctx, userID)
	}
}

// --- жёсткая параллель на множество пользователей и групп ---
func BenchmarkMemoryStorage_HeavyParallel_Mixed(b *testing.B) {
	store := NewMemoryStorage(30*time.Second, 30*time.Minute)
	ctx := context.Background()

	const users = 128
	const groups = 64
	// прогрев: создадим структуры
	for u := 0; u < users; u++ {
		uid := int64(u + 1)
		for g := 0; g < groups; g++ {
			store.SetMedia(ctx, uid, "g"+strconv.Itoa(g), mf("photo", "warmup"))
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var i int
		for pb.Next() {
			uid := int64((i % users) + 1)
			gid := "g" + strconv.Itoa(i%groups)
			if i&3 == 0 {
				store.Set(ctx, uid, "k", i)
			} else if i&3 == 1 {
				store.Get(ctx, uid, "k")
			} else if i&3 == 2 {
				store.SetMedia(ctx, uid, gid, mf("photo", "p"))
			} else {
				store.GetMedia(ctx, uid, gid)
			}
			i++
		}
	})
}

// --- проверка, что CleanCache не ломает параллельные Set/Get ---

func BenchmarkMemoryStorage_CleanCache_WithConcurrentTraffic(b *testing.B) {
	store := NewMemoryStorage(30*time.Second, 30*time.Minute)
	ctx := context.Background()
	const users = 64

	// фоновые писатели/читатели
	var wg sync.WaitGroup
	stop := make(chan struct{})
	for w := 0; w < 8; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			var i int
			for {
				select {
				case <-stop:
					return
				default:
					uid := int64((i % users) + 1)
					if i&1 == 0 {
						store.Set(ctx, uid, "k", i)
						store.SetMedia(ctx, uid, "g", mf("photo", "x"))
					} else {
						store.Get(ctx, uid, "k")
						store.GetMedia(ctx, uid, "g")
					}
					i++
				}
			}
		}(w)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		uid := int64((i % users) + 1)
		store.CleanCache(ctx, uid)
	}
	b.StopTimer()
	close(stop)
	wg.Wait()
}
