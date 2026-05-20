package dataselect

import (
	corev1 "k8s.io/api/core/v1"
	"k8soperation/global"
	"sort"
	"strings"
	"time"
)

// DataCell 定义接口
// GetCreation 获取创建时间
// GetName 获取对象名称
type DataCell interface {
	GetCreation() time.Time
	GetName() string
}

// DataSelector 用于封装排除、过滤以及分页的数据类型
type DataSelector struct {
	GenericDataList []DataCell
	DataSelect      *DataSelectQuery
}

// DataSelectQuery 定义过滤和分页的结构体
// 过滤使用的 Name
// 分页使用的 Limit 和 Page
// Limit 是单页数据条数
// Page 是第几页
type DataSelectQuery struct {
	Filter   *FilterQuery
	Paginate *PaginateQuery
}

type FilterQuery struct {
	Name string
}

type PaginateQuery struct {
	Limit int
	Page  int
}

// 实现自定义结构的排序
// 需要重写 Len(),Swap(),Less()

// Len 用户获取数据的长度
func (d *DataSelector) Len() int {
	return len(d.GenericDataList)
}

// Swap 用于数据比较大小之后的位置变更
// i,j 数组下标
func (d *DataSelector) Swap(i, j int) {
	d.GenericDataList[i], d.GenericDataList[j] = d.GenericDataList[j], d.GenericDataList[i]
}

// Less 用于比较大小
// i,j 数组下标
func (d *DataSelector) Less(i, j int) bool {
	a := d.GenericDataList[i].GetCreation()
	b := d.GenericDataList[j].GetCreation()
	return b.Before(a)
}

// Sort 用于触发排序
func (d *DataSelector) Sort() *DataSelector {
	sort.Sort(d)
	return d
}

// Filter 用于过滤数据
// 比较数据的 Name 属性，若包含，则返回
func (d *DataSelector) Filter() *DataSelector {
	if d.DataSelect == nil || d.DataSelect.Filter == nil {
		return d
	}

	// 取过滤关键字，去掉首尾空格并转小写
	kw := strings.ToLower(strings.TrimSpace(d.DataSelect.Filter.Name))
	if kw == "" {
		// 关键字为空，直接返回全部数据
		return d
	}

	var filtered []DataCell
	for _, item := range d.GenericDataList {
		// 名称同样转小写，保证大小写不敏感
		name := strings.ToLower(item.GetName())
		// 检查名称中是否包含关键字
		if strings.Contains(name, kw) {
			// 如果包含关键字，则将该项添加到过滤后的列表中
			filtered = append(filtered, item)
		}
	}

	// 将过滤后的列表赋值给通用数据列表属性
	d.GenericDataList = filtered
	// 返回处理后的数据对象
	return d
}

// Paginate 用于数组分页，根据Limit和 Page传参，返回数据
func (d *DataSelector) Paginate() *DataSelector {
	// 获取分页参数
	// limit 表示每页记录数
	limit := d.DataSelect.Paginate.Limit
	// page 表示当前页码
	page := d.DataSelect.Paginate.Page

	// 检验参数合法性
	if limit <= 0 || page <= 0 {
		return d
	}

	// 定义取数范围
	// 计算分页的起始索引和结束索引
	// startIndex: 当前页的起始位置，计算公式为 每页条数 * (当前页码 - 1)
	// endIndex: 当前页的结束位置，计算公式为 每页条数 * 当前页码
	startIndex := limit * (page - 1)
	endIndex := limit * page

	// 检查结束索引是否超出数据列表总长度
	// 如果超出，则将结束索引设置为数据列表的总长度，避免数组越界
	if endIndex > len(d.GenericDataList) {
		endIndex = len(d.GenericDataList)
	}

	// 根据计算出的起始索引和结束索引，对数据列表进行切片处理
	// 只保留当前页的数据，其他数据被舍弃
	d.GenericDataList = d.GenericDataList[startIndex:endIndex]
	// 返回处理后的数据对象
	return d
}

// NewDataSelector 是一个创建并返回新的DataSelector实例的函数
// 它接受四个参数：
//
//	cells: DataCell类型的切片，用于存储数据单元
//	name: 字符串类型，用于设置过滤器的名称
//	limit: 整型，用于设置分页查询的每页数据限制数量
//	page: 整型，用于设置分页查询的页码
//
// 函数返回一个指向DataSelector结构体的指针
func NewDataSelector(cells []DataCell, name string, limit, page int) *DataSelector {
	return &DataSelector{ // 返回一个新的DataSelector实例的指针
		GenericDataList: cells, // 将传入的数据单元切片赋值给GenericDataList字段
		DataSelect: &DataSelectQuery{ // 初始化DataSelectQuery指针
			Filter: &FilterQuery{Name: name}, // 初始化FilterQuery指针，并设置名称
			Paginate: &PaginateQuery{ // 初始化PaginateQuery指针
				Limit: limit, // 设置每页数据限制数量
				Page:  page,  // 设置当前页码
			},
		},
	}
}

func NewPodLogOptions(client *k8s.Client, container string, tail int64, follow bool) *corev1.PodLogOptions {
	// 1) 归一化 tail
	if tail < 0 { // 检查 tail 是否为负数
		tail = 0 // 如果 tail 值为负数，则重置为 0
	}

	// 想要“tail=0 => 不强制回放（让 kube 走默认/全量）”
	// 则仅当 tail>0 才设置 TailLines；否则置为 nil
	useTail := tail

	// 可选：为 follow 且未指定 tail 的情况，给一个较小的默认，避免一次性回放太多
	// 判断是否需要设置默认的tail行数
	// 条件：开启了follow模式、未手动指定tail行数、且配置中设置了默认tail行数
	if follow && useTail == 0 && client.PodLogSetting.TailDefault > 0 {
		// 设置tail行数为配置中的默认值
		useTail = client.PodLogSetting.TailDefault
	}

	// 上限
	// 如果全局Pod日志设置中的TailMax大于0，并且当前使用的useTail值大于TailMax
	// 则将 useTail的值限制为TailMax的最大值，确保不超过系统允许的最大日志行数
	if tailMax := client.PodLogSetting.TailMax; tailMax > 0 && useTail > tailMax {
		useTail = tailMax
	}

	// 2) 组装选项
	// 创建Pod日志选项的结构体指针，用于配置获取Pod日志的参数
	opts := &corev1.PodLogOptions{
		Timestamps: client.PodLogSetting.Timestamps, // 是否显示时间戳
		Previous:   client.PodLogSetting.Previous,   // 是否获取前一个容器的日志
		Follow:     follow,                          // 是否实时跟踪日志
	}

	// 容器名非空再赋值，避免发空串
	// 只有当容器名称不为空时才设置容器参数，防止发送空字符串导致请求错误
	if container != "" {
		opts.Container = container
	}

	// 只有 >0 才设置 TailLines；=0 则留 nil → 由 appserver 决定（可能是全量）
	if useTail > 0 {
		// 注意：取地址没问题，编译器会做逃逸
		opts.TailLines = &useTail
	}

	// Follow 时不建议 LimitBytes，避免截断；一次性模式可按需限制返回体大小
	if !follow {
		if lb := client.PodLogSetting.LimitBytes; lb > 0 {
			opts.LimitBytes = &lb
		}
	}

	return opts
}

// TotalCount 返回过滤后的总数（分页前的数量）
func (d *DataSelector) TotalCount() int {
	// 注意：这里直接返回当前列表长度即可，
	// 一般在调用方应该是先 Filter() 再调 TotalCount()
	return len(d.GenericDataList)
}
