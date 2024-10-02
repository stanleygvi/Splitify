import random
import string

def calc_slices(length):
    return (length + 99) // 100 if length > 0 else 0

def convert_to_string(index, track):
    artists = ', '.join(artist['name'] for artist in track['artists'])
    return f"{{{index}: {track['name']}, {artists}}},"

def generate_random_string(length: int)->str:

    letters = string.ascii_lowercase
    result_str = ''.join(random.choice(letters) for i in range(length))
    return result_str
    