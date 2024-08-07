package config

import (
	"fmt"
	"regexp"
	"strconv"
)

// Common fields across all Parameter instances
type ParameterCommon struct {
	// The parameter's ID, used for serialization and deserialization
	ID string

	// The parameter's human-readable name
	Name string

	// A description of this parameter / setting
	Description string

	// The max length of the parameter, in characters, if it's free-form input
	MaxLength int

	// An optional regex used to validate free-form input for the parameter
	Regex string

	// True if this is an advanced parameter and should be hidden unless advanced configuration mode is enabled
	Advanced bool

	// The list of Docker containers affected by changing this parameter
	// (these containers will require a restart for the change to take effect)
	AffectsContainers []ContainerID

	// A list of Docker container environment variables that should be set to this parameter's value
	EnvironmentVariables []string

	// Whether or not the parameter is allowed to be blank
	CanBeBlank bool

	// True to reset the parameter's value to the default option after the config is updated
	OverwriteOnUpgrade bool

	// Descriptions of the parameter that change depending on the selected network
	DescriptionsByNetwork map[Network]string
}

// Set the network-specific description of the parameter
func (p *ParameterCommon) UpdateDescription(network Network) {
	if p.DescriptionsByNetwork != nil {
		newDesc, exists := p.DescriptionsByNetwork[network]
		if exists {
			p.Description = newDesc
		}
	}
}

// A parameter that can be configured by the user
type Parameter[Type comparable] struct {
	*ParameterCommon
	Default map[Network]Type
	Value   Type
	Options []*ParameterOption[Type]
}

// An interface for typed Parameter structs, to get common fields from them
type IParameter interface {
	// Get the parameter's common fields
	GetCommon() *ParameterCommon

	// Get the common fields from each ParameterOption (returns nil if this isn't a choice parameter)
	GetOptions() []IParameterOption

	// Set the parameter to the default value
	SetToDefault(network Network)

	// Get the parameter's value
	GetValueAsAny() any

	// Get the parameter's value as a string
	String() string

	// Get the parameter's default value for the supplied network as a string
	GetDefaultAsAny(network Network) any

	// Deserializes a string into this parameter's value
	Deserialize(serializedParam string, network Network) error

	// Set the parameter's value explicitly; panics if it's the wrong type
	SetValue(value any)

	// Sets the default value for the provided network
	SetDefaultValueForNetwork(serializedDefault string, network Network) error

	// Change the current network
	ChangeNetwork(oldNetwork Network, newNetwork Network)
}

// Get the parameter's common fields
func (p *Parameter[_]) GetCommon() *ParameterCommon {
	return p.ParameterCommon
}

// Get the common fields from each ParameterOption (returns nil if this isn't a choice parameter)
func (p *Parameter[_]) GetOptions() []IParameterOption {
	if len(p.Options) == 0 {
		return nil
	}
	opts := make([]IParameterOption, len(p.Options))
	for i, param := range p.Options {
		opts[i] = param
	}
	return opts
}

// Set the value to the default for the provided config's network
func (p *Parameter[Type]) SetToDefault(network Network) {
	p.Value = p.GetDefault(network)
}

// Get the default value for the provided network
func (p *Parameter[Type]) GetDefault(network Network) Type {
	defaultSetting, exists := p.Default[network]
	if !exists {
		defaultSetting, exists = p.Default[Network_All]
		if !exists {
			panic(fmt.Sprintf("parameter [%s] doesn't have a default for network %s or all networks", p.Name, network))
		}
	}

	return defaultSetting
}

// Get the parameter's value
func (p *Parameter[_]) GetValueAsAny() any {
	return p.Value
}

// Get the parameter's value as a string
func (p *Parameter[_]) String() string {
	return fmt.Sprint(p.Value)
}

// Get the default value for the provided network
func (p *Parameter[_]) GetDefaultAsAny(network Network) any {
	return p.GetDefault(network)
}

// Deserializes a string into this parameter's value
func (p *Parameter[_]) Deserialize(serializedParam string, network Network) error {
	value, err := p.deserializeToType(serializedParam, network)
	if err != nil {
		// Set to the default option if none of the options match
		if len(p.Options) > 0 {
			p.SetToDefault(network)
			return nil
		}

		// Set blanks to the default if blank isn't allowed
		if !p.CanBeBlank && serializedParam == "" {
			p.SetToDefault(network)
			return nil
		}

		return fmt.Errorf("cannot deserialize parameter [%s]: %w", p.ID, err)
	}

	p.Value = value
	return nil
}

// Set the parameter's value
func (p *Parameter[Type]) SetValue(value any) {
	typedVal, ok := value.(Type)
	if !ok {
		panic(fmt.Sprintf("attempted to set param [%s] to [%v] but it was the wrong type", p.Name, value))
	}
	p.Value = typedVal
}

// Apply a network change to a parameter
func (p *Parameter[_]) ChangeNetwork(oldNetwork Network, newNetwork Network) {

	// Get the current value and the defaults per-network
	currentValue := p.Value
	oldDefault, exists := p.Default[oldNetwork]
	if !exists {
		oldDefault = p.Default[Network_All]
	}
	newDefault, exists := p.Default[newNetwork]
	if !exists {
		newDefault = p.Default[Network_All]
	}

	// If the old value matches the old default, replace it with the new default
	if currentValue == oldDefault {
		p.Value = newDefault
	}

	// Update the description, if applicable
	p.UpdateDescription(newNetwork)
}

// Sets the default value for the provided network
func (p *Parameter[Type]) SetDefaultValueForNetwork(serializedDefault string, network Network) error {
	value, err := p.deserializeToType(serializedDefault, network)
	if err != nil {
		return fmt.Errorf("cannot set default of parameter [%s] for network [%v]: %w", p.ID, network, err)
	}

	p.Default[network] = value
	return nil
}

// Deserializes a string into this parameter's value type
func (p *Parameter[Type]) deserializeToType(serializedDefault string, network Network) (Type, error) {
	var defaultValue Type
	if len(p.Options) > 0 {
		for _, option := range p.Options {
			optionVal := option.String()
			if optionVal == serializedDefault {
				defaultValue = option.Value
				return defaultValue, nil
			}
		}
		return defaultValue, fmt.Errorf("value [%s] is not a valid option", serializedDefault)
	}

	// Set blanks to the default if blank isn't allowed
	if !p.CanBeBlank && serializedDefault == "" {
		return defaultValue, fmt.Errorf("value cannot be blank")
	}

	var err error
	switch value := any(&defaultValue).(type) {
	case *int64:
		*value, err = strconv.ParseInt(serializedDefault, 0, 0)
	case *uint64:
		*value, err = strconv.ParseUint(serializedDefault, 0, 0)
	case *uint16:
		var result uint64
		result, err = strconv.ParseUint(serializedDefault, 0, 16)
		*value = uint16(result)
	case *bool:
		*value, err = strconv.ParseBool(serializedDefault)
	case *float64:
		*value, err = strconv.ParseFloat(serializedDefault, 64)
	case *string:
		if p.CanBeBlank && serializedDefault == "" {
			*value = ""
		} else if p.MaxLength > 0 && len(serializedDefault) > p.MaxLength {
			err = fmt.Errorf("value [%s] is longer than the max length of [%d]", serializedDefault, p.MaxLength)
		} else if p.Regex != "" {
			regex := regexp.MustCompile(p.Regex)
			if !regex.MatchString(serializedDefault) {
				err = fmt.Errorf("value [%s] did not match the expected format", serializedDefault)
			}
		} else {
			*value = serializedDefault
		}
	}
	return defaultValue, err
}
