package media

import (
	"sync"
	"testing"
	"time"
)

// helper для быстрого создания File
func f(tpe, id string) File {
	return File{Type: tpe, FileID: id}
}

func TestAddFileAndFilesIsolation(t *testing.T) {
	md := &MediaData{}
	md.AddFile(f("photo", "1"))
	md.AddFile(f("video", "2"))

	files := md.Files()
	if len(files) != 2 {
		t.Fatalf("ожидалось 2 файла, получили %d", len(files))
	}
	if files[0].FileID != "1" || files[1].FileID != "2" {
		t.Errorf("неправильные файлы: %+v", files)
	}

	// проверим изоляцию: меняем срез снаружи
	files[0].FileID = "xxx"
	if md.Files()[0].FileID == "xxx" {
		t.Error("Files() не вернул копию, внутренние данные повреждены")
	}
}

func TestTouchAndElapsed(t *testing.T) {
	md := &MediaData{}

	// сразу должно быть true, т.к. lastUpdate нулевой
	if !md.Elapsed(0) {
		t.Error("Elapsed при пустом времени должен быть true")
	}

	md.Touch()
	if md.Elapsed(time.Hour) {
		t.Error("сразу после Touch время не должно истечь")
	}

	time.Sleep(10 * time.Millisecond)
	if !md.Elapsed(0) {
		t.Error("после сна >0 должно быть истекшим для 0")
	}
}

func TestConcurrencySafety(t *testing.T) {
	md := &MediaData{}
	var wg sync.WaitGroup

	// параллельно добавляем файлы
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			md.AddFile(f("photo", string(rune('A'+i%26))))
		}(i)
	}

	// параллельно читаем
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = md.Files()
			_ = md.Elapsed(time.Second)
		}()
	}

	wg.Wait()

	if len(md.Files()) != 100 {
		t.Errorf("ожидалось 100 файлов, получили %d", len(md.Files()))
	}
}

func TestMixedOperations(t *testing.T) {
	md := &MediaData{}

	md.AddFile(f("photo", "x"))
	md.Touch()
	time.Sleep(1 * time.Millisecond)
	md.AddFile(f("video", "y"))

	files := md.Files()
	if len(files) != 2 {
		t.Fatalf("ожидалось 2 файла, получили %d", len(files))
	}

	if files[0].Type != "photo" || files[1].Type != "video" {
		t.Errorf("неверные типы: %+v", files)
	}

	// Touch обновляет lastUpdate, AddFile — нет
	md.Touch()
	if md.Elapsed(time.Hour) {
		t.Error("после повторного Touch время не должно истечь")
	}
}

func BenchmarkAddFile(b *testing.B) {
	md := &MediaData{}
	file := f("photo", "id")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		md.AddFile(file)
	}
}

func BenchmarkFiles(b *testing.B) {
	md := &MediaData{}
	for i := 0; i < 1000; i++ {
		md.AddFile(f("photo", string(rune(i))))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = md.Files()
	}
}

func BenchmarkAddFileParallel(b *testing.B) {
	md := &MediaData{}
	file := f("video", "id")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			md.AddFile(file)
		}
	})
}

func BenchmarkFilesParallel(b *testing.B) {
	md := &MediaData{}
	for i := 0; i < 10000; i++ {
		md.AddFile(f("video", string(rune(i))))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = md.Files()
		}
	})
}

// BenchmarkTouch - проверяет стоимость обновления lastUpdate.
func BenchmarkTouch(b *testing.B) {
	md := &MediaData{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		md.Touch()
	}
}

// BenchmarkElapsedShort - проверка Elapsed при маленьком t.
func BenchmarkElapsedShort(b *testing.B) {
	md := &MediaData{}
	md.Touch()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = md.Elapsed(time.Microsecond)
	}
}

// BenchmarkElapsedLong - проверка Elapsed при большом t.
func BenchmarkElapsedLong(b *testing.B) {
	md := &MediaData{}
	md.Touch()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = md.Elapsed(time.Hour)
	}
}

// BenchmarkFilesSmall - копирование короткого среза.
func BenchmarkFilesSmall(b *testing.B) {
	md := &MediaData{}
	for i := 0; i < 10; i++ {
		md.AddFile(f("photo", string(rune(i))))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = md.Files()
	}
}

// BenchmarkFilesLarge - копирование очень длинного среза.
func BenchmarkFilesLarge(b *testing.B) {
	md := &MediaData{}
	for i := 0; i < 100_000; i++ {
		md.AddFile(f("video", string(rune(i%1000))))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = md.Files()
	}
}

// BenchmarkMixedParallel - одновременно AddFile и Files в параллели.
func BenchmarkMixedParallel(b *testing.B) {
	md := &MediaData{}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%2 == 0 {
				md.AddFile(f("photo", "id"))
			} else {
				_ = md.Files()
			}
			i++
		}
	})
}
