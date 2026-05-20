package setting

import (
	"encoding/json"
	"k8soperation/pkg/logger"
	"k8soperation/pkg/setting/types"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const configEnvKey = "APP_CONFIG"

var (
	AppConfig *types.AppConfig
	log       = logger.NewBootstrapLogger(logger.AddCaller(), logger.AddCallerSkip(1))
)

func SetUpAppConfig(configFile string) {
	config := types.AppConfig{}
	if err := LoadAppConfig(&config, configFile); err != nil {
		log.Fatalf("%+v", errors.WithMessage(err, "setup failed"))
	}
	AppConfig = &config
}

func LoadAppConfig(config *types.AppConfig, configFile string) error {
	return setup(config, configFile)
}

func setup(config interface{}, configFile string) error {
	if configFile == "" {
		configFile = strings.TrimSpace(os.Getenv(configEnvKey))
		if configFile == "" {
			viper.SetConfigName("config")
			viper.AddConfigPath("configs")
		}
	}
	if filepath.IsAbs(configFile) || filepath.Dir(configFile) != "." {
		viper.SetConfigFile(configFile)
	}

	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		return errors.Wrapf(err, "read config %s failed", configFile)
	}
	if err := viper.Unmarshal(config); err != nil {
		return errors.Wrapf(err, "unmarshal config %s failed", configFile)
	}
	setDefaults(config)
	return nil
}

func setDefaults(p interface{}) {
	val := reflect.ValueOf(p).Elem()
	typ := reflect.TypeOf(p).Elem()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		tag := typ.Field(i).Tag.Get("default")
		if field.Kind() == reflect.Struct {
			setDefaults(field.Addr().Interface())
		} else if field.Kind() == reflect.Slice {
			// 处理切片赋默认值
			if tag != "" && (field.IsZero() || field.Len() == 0) {
				tempSlice := make([]interface{}, 0)
				err := json.Unmarshal([]byte(tag), &tempSlice)
				if err != nil {
					log.Fatalf("%+v", errors.Wrapf(err, "Unmarshal to []string failed "))
				}
				// 根据目标类型创建具体类型的切片
				targetType := field.Type().Elem() // 获取切片元素类型
				resultSlice := reflect.MakeSlice(field.Type(), len(tempSlice), len(tempSlice))
				// 类型转换赋值
				for i, item := range tempSlice {
					val := reflect.ValueOf(item)
					// 处理类型不匹配的情况（如JSON数字转float64，但目标是int）
					if !val.Type().ConvertibleTo(targetType) {
						// 尝试通过json重新序列化/反序列化转换类型
						jsonData, _ := json.Marshal(item)
						newVal := reflect.New(targetType)
						if err := json.Unmarshal(jsonData, newVal.Interface()); err == nil {
							val = newVal.Elem()
						}
					}
					if val.Type().ConvertibleTo(targetType) {
						resultSlice.Index(i).Set(val.Convert(targetType))
					} else {
						log.Errorf("cannot convert %v to %s", val.Type(), targetType)
					}
				}
				field.Set(resultSlice)
			}

		} else if field.Interface() == reflect.Zero(field.Type()).Interface() && tag != "" {
			switch field.Kind() {
			case reflect.String:
				field.SetString(tag)
			case reflect.Int, reflect.Int32, reflect.Int64, reflect.Uint8, reflect.Uint32, reflect.Uint64:
				defaultVal, err := strconv.Atoi(tag)
				if err != nil {
					log.Fatalf("%+v", errors.Wrapf(err, "convert %v to int/int32/int64/uin8/uint32/uin64 failed", tag))
				}
				field.SetInt(int64(defaultVal))
			case reflect.Float32, reflect.Float64:
				defaultVal, err := strconv.ParseFloat(tag, 64)
				if err != nil {
					log.Fatalf("%+v", errors.Wrapf(err, "convert %v to float32/float64 failed", tag))
				}
				field.SetFloat(defaultVal)
			case reflect.Bool:
				defaultVal, err := strconv.ParseBool(tag)
				if err != nil {
					log.Fatalf("%+v", errors.Wrapf(err, "convert %v to bool failed", tag))
				}
				field.SetBool(defaultVal)
			}
		}
	}
}
