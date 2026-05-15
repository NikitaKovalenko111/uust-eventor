import * as ImagePicker from 'expo-image-picker';
import { NativeStackScreenProps } from '@react-navigation/native-stack';
import { useEffect, useState } from 'react';
import { Alert, Image, StyleSheet, Text, View } from 'react-native';
import { AppHeader } from '../components/AppHeader';
import { EventCard } from '../components/EventCard';
import { FormInput } from '../components/FormInput';
import { PrimaryButton } from '../components/PrimaryButton';
import { ScreenContainer } from '../components/ScreenContainer';
import { theme } from '../constants/theme';
import { uploadEventImageApi } from '../api/api';
import { useAppDispatch, useAppSelector } from '../redux/hooks';
import { createEventRequest, fetchEventsRequest, registerForEventRequest, setSearchText } from '../redux/slices/eventsSlice';
import { RootStackParamList } from '../navigation/types';

type Props = NativeStackScreenProps<RootStackParamList, 'Events'>;

export const EventsPage = ({ navigation }: Props) => {
  const dispatch = useAppDispatch();
  const { list, searchText } = useAppSelector((state) => state.events);
  const isAuthenticated = useAppSelector((state) => state.auth.isAuthenticated);
  const userRole = useAppSelector((state) => state.auth.user?.role);
  const userId = useAppSelector((state) => state.auth.user?.id ?? '');
  const isModerator = userRole === 'moderator';

  const [showCreateForm, setShowCreateForm] = useState(false);
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [dateValue, setDateValue] = useState('');
  const [location, setLocation] = useState('');
  const [tagsValue, setTagsValue] = useState('');
  const [coverImageId, setCoverImageId] = useState('');
  const [coverImageUri, setCoverImageUri] = useState('');

  useEffect(() => {
    dispatch(fetchEventsRequest());
  }, [dispatch]);

  const pickEventCover = async () => {
    const permission = await ImagePicker.requestMediaLibraryPermissionsAsync();
    if (!permission.granted) {
      Alert.alert('Нет доступа', 'Разрешите доступ к фото для выбора обложки');
      return;
    }

    const result = await ImagePicker.launchImageLibraryAsync({
      mediaTypes: ['images'],
      allowsEditing: true,
      quality: 0.8,
    });

    if (result.canceled || !result.assets[0]?.uri) {
      return;
    }

    setCoverImageUri(result.assets[0].uri);

    try {
      const uploaded = await uploadEventImageApi(result.assets[0].uri);
      setCoverImageId(uploaded.image_id);
    } catch (error) {
      Alert.alert('Не удалось загрузить обложку', error instanceof Error ? error.message : 'Попробуйте еще раз');
      setCoverImageId('');
    }
  };

  const createEvent = () => {
    if (!title || !description || !dateValue || !location) {
      Alert.alert('Заполните поля', 'Укажите название, описание, дату и место мероприятия');
      return;
    }

    dispatch(
      createEventRequest({
        title,
        description,
        date: dateValue,
        location,
        tags: tagsValue.split(',').map((item) => item.trim()).filter(Boolean),
        imageUri: coverImageId,
      })
    );

    setTitle('');
    setDescription('');
    setDateValue('');
    setLocation('');
    setTagsValue('');
    setCoverImageId('');
    setCoverImageUri('');
    setShowCreateForm(false);
  };

  return (
    <ScreenContainer>
      <AppHeader
        title="Мероприятия"
        subtitle="Поиск по всем полям"
        onActionPress={() => {
          if (!isAuthenticated) {
            navigation.navigate('Auth');
            return;
          }

          navigation.navigate(userRole === 'moderator' ? 'ModeratorCabinet' : 'UserCabinet');
        }}
        actionIcon={isAuthenticated ? 'person-circle-outline' : 'log-in-outline'}
      />

      {isModerator ? (
        <View style={styles.createSection}>
          <PrimaryButton
            title={showCreateForm ? 'Скрыть форму создания' : 'Создать мероприятие'}
            onPress={() => setShowCreateForm((prev) => !prev)}
          />

          {showCreateForm ? (
            <View style={styles.createCard}>
              <View style={styles.coverBlock}>
                {coverImageUri ? (
                  <Image source={{ uri: coverImageUri }} style={styles.coverPreview} resizeMode="cover" />
                ) : (
                  <Text style={styles.coverPlaceholder}>Обложка еще не загружена</Text>
                )}
                <PrimaryButton title="Загрузить обложку" type="outline" onPress={pickEventCover} />
              </View>
              <FormInput label="Название" value={title} onChangeText={setTitle} />
              <FormInput label="Описание" value={description} onChangeText={setDescription} multiline />
              <FormInput label="Дата" value={dateValue} onChangeText={setDateValue} placeholder="YYYY-MM-DD" />
              <FormInput label="Место" value={location} onChangeText={setLocation} />
              <FormInput label="Теги" value={tagsValue} onChangeText={setTagsValue} placeholder="Например: IT, AI, Карьера" />
              <PrimaryButton title="Опубликовать мероприятие" onPress={createEvent} />
            </View>
          ) : null}
        </View>
      ) : null}

      <View style={styles.searchCard}>
        <FormInput
          label="Поиск"
          value={searchText}
          onChangeText={(value) => {
            dispatch(setSearchText(value));
            dispatch(fetchEventsRequest());
          }}
          placeholder="Название, описание, место, теги"
        />
      </View>

      <Text style={styles.sectionTitle}>Найдено: {list.length}</Text>

      {list.map((eventItem) => (
        <EventCard
          key={eventItem.id}
          eventItem={eventItem}
          isRegistered={eventItem.attendees.includes(userId)}
          onPress={() => navigation.navigate('EventDetails', { eventId: eventItem.id })}
          onToggleRegistration={
            isAuthenticated && eventItem.creatorId !== userId
              ? () => dispatch(registerForEventRequest(eventItem.id))
              : undefined
          }
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
  createSection: {
    gap: theme.spacing.sm,
  },
  createCard: {
    borderWidth: 1,
    borderColor: theme.colors.border,
    borderRadius: theme.radius.md,
    padding: theme.spacing.md,
    backgroundColor: theme.colors.surface,
    gap: theme.spacing.sm,
  },
  coverBlock: {
    gap: theme.spacing.xs,
  },
  coverPreview: {
    width: '100%',
    height: 170,
    borderRadius: theme.radius.md,
    backgroundColor: theme.colors.card,
  },
  coverPlaceholder: {
    width: '100%',
    height: 170,
    borderRadius: theme.radius.md,
    backgroundColor: theme.colors.card,
    color: theme.colors.muted,
    textAlign: 'center',
    textAlignVertical: 'center',
    lineHeight: 170,
    overflow: 'hidden',
  },
  sectionTitle: {
    color: theme.colors.text,
    fontSize: 16,
    fontWeight: '700',
  },
});
