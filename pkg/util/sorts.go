package util

import "tilapia/models"

type SortMenuPerms []*models.MenuPerms

func (t SortMenuPerms) Len() int           { return len(t) }
func (t SortMenuPerms) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t SortMenuPerms) Less(i, j int) bool { return t[i].ID < t[j].ID }
