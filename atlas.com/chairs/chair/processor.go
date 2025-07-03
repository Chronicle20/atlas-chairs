package chair

import (
	_map2 "atlas-chairs/data/map"
	chair2 "atlas-chairs/kafka/message/chair"
	"atlas-chairs/kafka/producer"
	"context"
	"errors"
	"github.com/Chronicle20/atlas-constants/field"
	"github.com/sirupsen/logrus"
	"math"
)

type Processor interface {
	GetById(characterId uint32) (Model, error)
	Set(field field.Model, chairType string, chairId uint32, characterId uint32) error
	Clear(field field.Model, characterId uint32) error
}

type ProcessorImpl struct {
	l   logrus.FieldLogger
	ctx context.Context
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context) Processor {
	return &ProcessorImpl{
		l:   l,
		ctx: ctx,
	}
}

func (p *ProcessorImpl) GetById(characterId uint32) (Model, error) {
	m, ok := GetRegistry().Get(characterId)
	if !ok {
		return Model{}, errors.New("not found")
	}
	return m, nil
}

func (p *ProcessorImpl) Set(field field.Model, chairType string, chairId uint32, characterId uint32) error {
	p.l.Debugf("Attempting to sit in chair [%d] for character [%d].", chairId, characterId)
	_, err := p.GetById(characterId)
	if err == nil {
		p.l.Errorf("Character [%d] already sitting on chair.", characterId)
		_ = producer.ProviderImpl(p.l)(p.ctx)(chair2.EnvEventTopicStatus)(statusEventErrorProvider(field, chairType, chairId, characterId, chair2.ErrorTypeAlreadySitting))
		return errors.New("already sitting")
	}

	if chairType == chair2.ChairTypeFixed {
		var m _map2.Model
		m, err = _map2.NewProcessor(p.l, p.ctx).GetById(field.MapId())
		if err != nil {
			p.l.WithError(err).Errorf("Unable to retrieve map [%d] character [%d] is sitting in.", field.MapId(), characterId)
			_ = producer.ProviderImpl(p.l)(p.ctx)(chair2.EnvEventTopicStatus)(statusEventErrorProvider(field, chairType, chairId, characterId, chair2.ErrorTypeInternal))
			return err
		}

		if chairId >= m.Seats() {
			p.l.Errorf("Character [%d] is attempting to sit in fixed chair [%d] in map [%d], but that does not exist.", characterId, chairId, field.MapId())
			_ = producer.ProviderImpl(p.l)(p.ctx)(chair2.EnvEventTopicStatus)(statusEventErrorProvider(field, chairType, chairId, characterId, chair2.ErrorTypeDoesNotExist))
			return errors.New("chair does not exist")
		}

	}
	if chairType == chair2.ChairTypePortable {
		itemCategory := uint32(math.Floor(float64(chairId / 10000)))
		if itemCategory != 301 {
			p.l.Errorf("Character [%d] is attempting to sit in portable chair [%d] in map [%d], but that does not exist.", characterId, chairId, field.MapId())
			_ = producer.ProviderImpl(p.l)(p.ctx)(chair2.EnvEventTopicStatus)(statusEventErrorProvider(field, chairType, chairId, characterId, chair2.ErrorTypeDoesNotExist))
			return errors.New("chair does not exist")
		}

		// TODO ensure character has item.
	}

	c := Model{
		id:        chairId,
		chairType: chairType,
	}

	GetRegistry().Set(characterId, c)
	return producer.ProviderImpl(p.l)(p.ctx)(chair2.EnvEventTopicStatus)(statusEventUsedProvider(field, chairType, chairId, characterId))
}

func (p *ProcessorImpl) Clear(field field.Model, characterId uint32) error {
	p.l.Debugf("Attempting to clear for character [%d].", characterId)
	c, err := p.GetById(characterId)
	if err != nil {
		p.l.WithError(err).Errorf("Failed to get chair for character [%d].", characterId)
		return err
	}
	existed := GetRegistry().Clear(characterId)
	if existed {
		return producer.ProviderImpl(p.l)(p.ctx)(chair2.EnvEventTopicStatus)(statusEventCancelledProvider(field, c.Type(), c.Id(), characterId))
	}
	return nil
}
