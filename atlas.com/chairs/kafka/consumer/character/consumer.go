package character

import (
	"atlas-chairs/chair"
	consumer2 "atlas-chairs/kafka/consumer"
	"context"
	"github.com/Chronicle20/atlas-kafka/consumer"
	"github.com/Chronicle20/atlas-kafka/handler"
	"github.com/Chronicle20/atlas-kafka/message"
	"github.com/Chronicle20/atlas-kafka/topic"
	"github.com/Chronicle20/atlas-model/model"
	"github.com/sirupsen/logrus"
)

func InitConsumers(l logrus.FieldLogger) func(func(config consumer.Config, decorators ...model.Decorator[consumer.Config])) func(consumerGroupId string) {
	return func(rf func(config consumer.Config, decorators ...model.Decorator[consumer.Config])) func(consumerGroupId string) {
		return func(consumerGroupId string) {
			rf(consumer2.NewConfig(l)("character_status")(EnvEventTopicCharacterStatus)(consumerGroupId), consumer.SetHeaderParsers(consumer.SpanHeaderParser, consumer.TenantHeaderParser))
		}
	}
}

func InitHandlers(l logrus.FieldLogger) func(rf func(topic string, handler handler.Handler) (string, error)) {
	return func(rf func(topic string, handler handler.Handler) (string, error)) {
		var t string
		t, _ = topic.EnvProvider(l)(EnvEventTopicCharacterStatus)()
		_, _ = rf(t, message.AdaptHandler(message.PersistentConfig(handleStatusEventLogout)))
		_, _ = rf(t, message.AdaptHandler(message.PersistentConfig(handleStatusEventChannelChange)))
	}
}

func handleStatusEventChannelChange(l logrus.FieldLogger, ctx context.Context, e statusEvent[statusEventChannelChangedBody]) {
	if e.Type != EventCharacterStatusTypeChannelChanged {
		return
	}
	_ = chair.Clear(l)(ctx)(e.CharacterId, e.Body.MapId)
}

func handleStatusEventLogout(l logrus.FieldLogger, ctx context.Context, e statusEvent[statusEventLogoutBody]) {
	if e.Type != EventCharacterStatusTypeLogout {
		return
	}
	_ = chair.Clear(l)(ctx)(e.CharacterId, e.Body.MapId)
}
