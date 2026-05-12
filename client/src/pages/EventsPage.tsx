import { NativeStackScreenProps } from '@react-navigation/native-stack';
import { StyleSheet, Text, View } from 'react-native';
import { AppHeader } from '../components/AppHeader';
import { EventCard } from '../components/EventCard';
import { FormInput } from '../components/FormInput';
import { ScreenContainer } from '../components/ScreenContainer';
import { theme } from '../constants/theme';
import { useAppDispatch, useAppSelector } from '../redux/hooks';
import { registerForEventRequest, setSearchDate, setSearchText } from '../redux/slices/eventsSlice';
import { RootStackParamList } from '../navigation/types';

type Props = NativeStackScreenProps<RootStackParamList, 'Events'>;

export const EventsPage = ({ navigation }: Props) => {
  const dispatch = useAppDispatch();
  const { list, searchText, searchDate } = useAppSelector((state) => state.events);
  const userId = useAppSelector((state) => state.auth.user?.id ?? '');

  const text = searchText.trim().toLowerCase();
  const date = searchDate.trim();
  const filteredEvents = list.filter((eventItem) => {
    const searchable = [
      eventItem.title,
      eventItem.shortDescription,
      eventItem.description,
      eventItem.location,
      eventItem.date,
      ...eventItem.tags,
    ]
      .join(' ')
      .toLowerCase();

    const hasText = text ? searchable.includes(text) : true;
    const hasDate = date ? eventItem.date.includes(date) : true;
    return hasText && hasDate;
  });

  return (
    <ScreenContainer>
      <AppHeader title="Мероприятия" subtitle="Поиск по всем полям" onActionPress={() => navigation.goBack()} actionIcon="arrow-back" />

      <View style={styles.searchCard}>
        <FormInput
          label="Поиск"
          value={searchText}
          onChangeText={(value) => dispatch(setSearchText(value))}
          placeholder="Название, описание, место, теги"
        />
        <FormInput
          label="Поиск по дате"
          value={searchDate}
          onChangeText={(value) => dispatch(setSearchDate(value))}
          placeholder="YYYY-MM-DD"
        />
      </View>

      <Text style={styles.sectionTitle}>Найдено: {filteredEvents.length}</Text>

      {filteredEvents.map((eventItem) => (
        <EventCard
          key={eventItem.id}
          eventItem={eventItem}
          isRegistered={eventItem.attendees.includes(userId)}
          onPress={() => navigation.navigate('EventDetails', { eventId: eventItem.id })}
          onToggleRegistration={() => dispatch(registerForEventRequest(eventItem.id))}
        />
      ))}
    </ScreenContainer>
  );
};

const styles = StyleSheet.create({
  searchCard: {
    borderWidth: 1,
    borderColor: theme.colors.border,
    borderRadius: theme.radius.md,
    padding: theme.spacing.md,
    backgroundColor: theme.colors.surface,
    gap: theme.spacing.sm,
  },
  sectionTitle: {
    color: theme.colors.text,
    fontSize: 16,
    fontWeight: '700',
  },
});
