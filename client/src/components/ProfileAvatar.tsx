import { Ionicons } from '@expo/vector-icons';
import { useEffect, useState } from 'react';
import { Image, StyleSheet, View } from 'react-native';
import { theme } from '../constants/theme';

type Props = {
  uri?: string | null;
  size?: number;
};

export const ProfileAvatar = ({ uri, size = 104 }: Props) => {
  const [imageFailed, setImageFailed] = useState(false);
  const hasImage = typeof uri === 'string' && uri.trim().length > 0;

  useEffect(() => {
    setImageFailed(false);
  }, [uri]);

  return (
    <View style={[styles.wrap, { width: size, height: size, borderRadius: size / 2 }]}>
      {hasImage && !imageFailed ? (
        <Image
          source={{ uri }}
          style={[styles.image, { borderRadius: size / 2 }]}
          resizeMode="cover"
          onError={() => setImageFailed(true)}
        />
      ) : (
        <View style={[styles.placeholder, { borderRadius: size / 2 }]}>
          <Ionicons name="person" size={size * 0.46} color={theme.colors.primary} />
        </View>
      )}
    </View>
  );
};

const styles = StyleSheet.create({
  wrap: {
    padding: 3,
    borderWidth: 2,
    borderColor: theme.colors.primary,
    backgroundColor: theme.colors.surface,
  },
  image: {
    width: '100%',
    height: '100%',
  },
  placeholder: {
    width: '100%',
    height: '100%',
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: theme.colors.card,
  },
});
