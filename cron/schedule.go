package cron

type Schedule []*CronJob

func (cj *Schedule) Len() int{
	return len(*cj)
}

func (cj *Schedule) Less(i, j int) bool {
	c := *cj
	return c[i].NextRunTime().Unix() <= c[j].NextRunTime().Unix()
}

func (cj *Schedule) Swap(i, j int)  {
	c := *cj
	c[i], c[j] = c[j], c[i]
}

func (cj *Schedule) Push(x interface{}) {
	*cj = append(*cj, x.(*CronJob))
}

func (cj *Schedule) Pop() interface{} {
	size := cj.Len()
	x := (*cj)[size-1]
	*cj = (*cj)[:size-1]
	return x
}

